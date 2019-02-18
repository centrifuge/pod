package txv1

import (
	"context"
	"fmt"
	"time"

	"github.com/centrifuge/go-centrifuge/transactions"

	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/protobufs/gen/go/transactions"
	"github.com/centrifuge/go-centrifuge/utils"
)

// extendedManager exposes package specific functions.
type extendedManager interface {
	transactions.Manager

	// saveTransaction only exposed for testing within package.
	// DO NOT use this outside of the package, use ExecuteWithinTX to initiate a transaction with management.
	saveTransaction(tx *transactions.Transaction) error

	// createTransaction only exposed for testing within package.
	// DO NOT use this outside of the package, use ExecuteWithinTX to initiate a transaction with management.
	createTransaction(accountID identity.CentID, desc string) (*transactions.Transaction, error)
}

// NewManager returns a Manager implementation.
func NewManager(config transactions.Config, repo transactions.Repository) transactions.Manager {
	return &manager{config: config, repo: repo}
}

// manager implements Manager.
// TODO [TXManager] convert this into an implementation of node.Server and start it at node start so that we can bring down transaction go routines cleanly
type manager struct {
	config transactions.Config
	repo   transactions.Repository
}

func (s *manager) GetDefaultTaskTimeout() time.Duration {
	return s.config.GetEthereumContextWaitTimeout()
}

func (s *manager) UpdateTaskStatus(accountID identity.CentID, id transactions.TxID, status transactions.Status, taskName, message string) error {
	tx, err := s.GetTransaction(accountID, id)
	if err != nil {
		return err
	}

	// status particular to the task
	tx.TaskStatus[taskName] = status
	tx.Logs = append(tx.Logs, transactions.NewLog(taskName, message))
	return s.saveTransaction(tx)
}

// ExecuteWithinTX executes a transaction within a transaction.
func (s *manager) ExecuteWithinTX(ctx context.Context, accountID identity.CentID, existingTxID transactions.TxID, desc string, work func(accountID identity.CentID, txID transactions.TxID, txMan transactions.Manager, err chan<- error)) (txID transactions.TxID, done chan bool, err error) {
	t, err := s.repo.Get(accountID, existingTxID)
	if err != nil {
		t = transactions.NewTransaction(accountID, desc)
		err := s.saveTransaction(t)
		if err != nil {
			return transactions.NilTxID(), nil, err
		}
	}
	done = make(chan bool)
	go func(ctx context.Context) {
		err := make(chan error)
		go work(accountID, t.ID, s, err)

		select {
		case e := <-err:
			tempTx, err := s.repo.Get(accountID, t.ID)
			if err != nil {
				log.Error(e, err)
				break
			}
			// update tx success status only if this wasn't an existing TX.
			// Otherwise it might update an existing tx pending status to success without actually being a success,
			// It is assumed that status update is already handled per task in that case.
			// Checking individual task success is upto the transaction manager users.
			if e == nil && transactions.TxIDEqual(existingTxID, transactions.NilTxID()) {
				tempTx.Status = transactions.Success
			} else if e != nil {
				tempTx.Status = transactions.Failed
			}
			e = s.saveTransaction(tempTx)
			if e != nil {
				log.Error(e)
			}
		case <-ctx.Done():
			msg := fmt.Sprintf("Transaction %s for account %s with description \"%s\" is stopped because of context close", t.ID.String(), t.CID, t.Description)
			log.Warningf(msg)
			tempTx, err := s.repo.Get(accountID, t.ID)
			if err != nil {
				log.Error(err)
				break
			}
			tempTx.Logs = append(tempTx.Logs, transactions.NewLog("context closed", msg))
			e := s.saveTransaction(tempTx)
			if e != nil {
				log.Error(e)
			}
		}
		done <- true
	}(ctx)
	return t.ID, done, nil
}

// saveTransaction saves the transaction.
func (s *manager) saveTransaction(tx *transactions.Transaction) error {
	err := s.repo.Save(tx)
	if err != nil {
		return err
	}
	return nil
}

// GetTransaction returns the transaction associated with identity and id.
func (s *manager) GetTransaction(accountID identity.CentID, id transactions.TxID) (*transactions.Transaction, error) {
	return s.repo.Get(accountID, id)
}

// createTransaction creates a new transaction and saves it to the DB.
func (s *manager) createTransaction(accountID identity.CentID, desc string) (*transactions.Transaction, error) {
	tx := transactions.NewTransaction(accountID, desc)
	return tx, s.saveTransaction(tx)
}

// WaitForTransaction blocks until transaction status is moved from pending state.
// Note: use it with caution as this will block.
func (s *manager) WaitForTransaction(accountID identity.CentID, txID transactions.TxID) error {
	// TODO change this to use a pre-saved done channel from ExecuteWithinTX, instead of a for loop, may require significant refactoring to handle the case of restarted node
	for {
		resp, err := s.GetTransactionStatus(accountID, txID)
		if err != nil {
			return err
		}

		switch transactions.Status(resp.Status) {
		case transactions.Failed:
			return errors.New("transaction failed: %v", resp.Message)
		case transactions.Success:
			return nil
		default:
			time.Sleep(10 * time.Millisecond)
			continue
		}
	}
}

// GetTransactionStatus returns the transaction status associated with identity and id.
func (s *manager) GetTransactionStatus(accountID identity.CentID, id transactions.TxID) (*transactionspb.TransactionStatusResponse, error) {
	tx, err := s.GetTransaction(accountID, id)
	if err != nil {
		return nil, err
	}

	var msg string
	lastUpdated := tx.CreatedAt.UTC()
	if len(tx.Logs) > 0 {
		log := tx.Logs[len(tx.Logs)-1]
		msg = log.Message
		lastUpdated = log.CreatedAt.UTC()
	}

	return &transactionspb.TransactionStatusResponse{
		TransactionId: tx.ID.String(),
		Status:        string(tx.Status),
		Message:       msg,
		LastUpdated:   utils.ToTimestamp(lastUpdated),
	}, nil
}
