package transactions

import (
	"time"

	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/protobufs/gen/go/transactions"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/satori/go.uuid"
)

// Service wraps the repository and exposes specific functions.
type Service interface {
	CreateTransaction(tenantID common.Address, desc string) (*Transaction, error)
	GetTransactionStatus(identity common.Address, id uuid.UUID) (*transactionspb.TransactionStatusResponse, error)
	WaitForTransaction(tenantID common.Address, txID uuid.UUID) error
}

// NewService returns a Service implementation.
func NewService(repo Repository) Service {
	return service{repo: repo}
}

// service implements Service.
type service struct {
	repo Repository
}

// CreateTransaction creates a new transaction and saves it to the DB.
func (s service) CreateTransaction(tenantID common.Address, desc string) (*Transaction, error) {
	tx := NewTransaction(tenantID, desc)
	return tx, s.repo.Save(tx)
}

// WaitForTransaction blocks until transaction status is moved from pending state.
// Note: use it with caution as this will block.
func (s service) WaitForTransaction(tenantID common.Address, txID uuid.UUID) error {
	for {
		resp, err := s.GetTransactionStatus(tenantID, txID)
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
func (s service) GetTransactionStatus(identity common.Address, id uuid.UUID) (*transactionspb.TransactionStatusResponse, error) {
	tx, err := s.repo.Get(identity, id)
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
