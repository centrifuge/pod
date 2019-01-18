package transactions

import (
	"sync"
	"time"

	"github.com/centrifuge/go-centrifuge/identity"

	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/protobufs/gen/go/transactions"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/satori/go.uuid"
)

// Service wraps the repository and exposes specific functions.
type Service interface {
	CreateTransaction(accountID identity.CentID, desc string) (*Transaction, error)
	GetTransaction(accountID identity.CentID, id uuid.UUID) (*Transaction, error)
	SaveTransaction(tx *Transaction) error
	GetTransactionStatus(identity identity.CentID, id uuid.UUID) (*transactionspb.TransactionStatusResponse, error)
	WaitForTransaction(accountID identity.CentID, txID uuid.UUID) error
	RegisterHandler(txID uuid.UUID, onUpdate func(status Status) error)
}

// NewService returns a Service implementation.
func NewService(repo Repository) Service {
	return &service{repo: repo, handlers: make(map[uuid.UUID][]func(status Status) error)}
}

// service implements Service.
type service struct {
	repo     Repository
	handlers map[uuid.UUID][]func(status Status) error
	mu       sync.Mutex
}

// SaveTransaction saves the transaction.
func (s *service) SaveTransaction(tx *Transaction) error {
	err := s.repo.Save(tx)
	if err != nil {
		return err
	}

	if len(tx.Logs) < 1 {
		return nil
	}

	s.mu.Lock()

	for _, h := range s.handlers[tx.ID] {
		err = h(tx.Status)
		if err != nil {
			break
		}
	}
	delete(s.handlers, tx.ID)
	s.mu.Unlock()

	// add the handler error as log and save it again
	if err != nil {
		log := tx.Logs[len(tx.Logs)-1]
		tx.Logs = append(tx.Logs, NewLog(log.Action, err.Error()))
		tx.Status = Failed
		return s.SaveTransaction(tx)
	}

	return nil
}

// GetTransaction returns the transaction associated with identity and id.
func (s *service) GetTransaction(accountID identity.CentID, id uuid.UUID) (*Transaction, error) {
	return s.repo.Get(accountID, id)
}

// CreateTransaction creates a new transaction and saves it to the DB.
func (s *service) CreateTransaction(accountID identity.CentID, desc string) (*Transaction, error) {
	tx := NewTransaction(accountID, desc)
	return tx, s.SaveTransaction(tx)
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

// RegisterHandler registers the handler to be triggered on transaction update.
// Handler is removed once triggered.
func (s *service) RegisterHandler(txID uuid.UUID, onUpdate func(status Status) error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.handlers[txID] = append(s.handlers[txID], onUpdate)
}
