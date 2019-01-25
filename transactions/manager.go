package transactions

import (
	"time"

	"github.com/centrifuge/go-centrifuge/identity"

	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/protobufs/gen/go/transactions"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/satori/go.uuid"
)

// Manager wraps the repository and exposes specific functions.
type Manager interface {
	ExecuteWithinTX(accountID identity.CentID, existingTxID uuid.UUID, desc string, work func(accountID identity.CentID, txID uuid.UUID, txMan Manager) error) (txID uuid.UUID, err error)
	CreateTransaction(accountID identity.CentID, desc string) (*Transaction, error)
	GetTransaction(accountID identity.CentID, id uuid.UUID) (*Transaction, error)
	saveTransaction(tx *Transaction) error
	GetTransactionStatus(accountID identity.CentID, id uuid.UUID) (*transactionspb.TransactionStatusResponse, error)
	WaitForTransaction(accountID identity.CentID, txID uuid.UUID) error
}

// NewManager returns a Manager implementation.
func NewManager(repo Repository) Manager {
	return &service{repo: repo}
}

// service implements Manager.
type service struct {
	repo Repository
}

func (s *service) ExecuteWithinTX(accountID identity.CentID, existingTxID uuid.UUID, desc string, work func(accountID identity.CentID, txID uuid.UUID, txMan Manager) error) (txID uuid.UUID, err error) {
	t, err := s.repo.Get(accountID, existingTxID)
	if err != nil {
		t = newTransaction(accountID, desc)
		err := s.saveTransaction(t)
		if err != nil {
			return uuid.Nil, err
		}
	}
	go func() {
		err = work(accountID, t.ID, s)
		if err != nil {
			t.Status = Failed
		}
		t.Status = Success
		err = s.saveTransaction(t)
		if err != nil {
			log.Error(err)
			return
		}
	}()
	return t.ID, nil
}

// saveTransaction saves the transaction.
func (s *service) saveTransaction(tx *Transaction) error {
	err := s.repo.Save(tx)
	if err != nil {
		return err
	}
	return nil
}

// GetTransaction returns the transaction associated with identity and id.
func (s *service) GetTransaction(accountID identity.CentID, id uuid.UUID) (*Transaction, error) {
	return s.repo.Get(accountID, id)
}

// CreateTransaction creates a new transaction and saves it to the DB.
func (s *service) CreateTransaction(accountID identity.CentID, desc string) (*Transaction, error) {
	tx := newTransaction(accountID, desc)
	return tx, s.saveTransaction(tx)
}

// WaitForTransaction blocks until transaction status is moved from pending state.
// Note: use it with caution as this will block.
func (s *service) WaitForTransaction(accountID identity.CentID, txID uuid.UUID) error {
	for {
		resp, err := s.GetTransactionStatus(accountID, txID)
		if err != nil {
			return err
		}

		switch Status(resp.Status) {
		case Failed:
			return errors.New("transaction failed: %v", resp.Message)
		case Success:
			return nil
		default:
			time.Sleep(10 * time.Millisecond)
			continue
		}
	}
}

// GetTransactionStatus returns the transaction status associated with identity and id.
func (s *service) GetTransactionStatus(identity identity.CentID, id uuid.UUID) (*transactionspb.TransactionStatusResponse, error) {
	tx, err := s.GetTransaction(identity, id)
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
