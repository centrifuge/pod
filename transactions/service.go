package transactions

import (
	"github.com/centrifuge/go-centrifuge/protobufs/gen/go/transactions"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/satori/go.uuid"
)

// Service wraps the repository and exposes specific functions.
type Service interface {
	GetTransactionStatus(identity common.Address, id uuid.UUID) (*transactionspb.TransactionStatusResponse, error)
}

// NewService returns a Service implementation.
func NewService(repo Repository) Service {
	return service{repo: repo}
}

// service implements Service.
type service struct {
	repo Repository
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
