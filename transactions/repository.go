package transactions

import (
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/storage"
	"github.com/satori/go.uuid"
)

const (
	// ErrTransactionMissing error when transaction doesn't exist in Repository.
	ErrTransactionMissing = errors.Error("transaction doesn't exist")

	// ErrKeyConstructionFailed error when the key construction failed.
	ErrKeyConstructionFailed = errors.Error("failed to construct transaction key")
)

// Repository can be implemented by a type that handles storage for transactions.
type Repository interface {
	Get(identity identity.CentID, id uuid.UUID) (*Transaction, error)
	Save(transaction *Transaction) error
}

// txRepository implements Repository.
type txRepository struct {
	repo storage.Repository
}

// NewRepository registers the the transaction model and returns the an implementation
// of the Repository.
func NewRepository(repo storage.Repository) Repository {
	repo.Register(&Transaction{})
	return &txRepository{repo: repo}
}

// getKey appends identity with id.
// With identity coming at first, we can even fetch transactions belonging to specific identity through prefix.
func getKey(cid identity.CentID, id uuid.UUID) ([]byte, error) {
	if uuid.Equal(uuid.Nil, id) {
		return nil, errors.New("transaction ID is not valid")
	}

	return append(cid[:], id.Bytes()...), nil
}

// Get returns the transaction associated with identity and id.
func (r *txRepository) Get(identity identity.CentID, id uuid.UUID) (*Transaction, error) {
	key, err := getKey(identity, id)
	if err != nil {
		return nil, errors.NewTypedError(ErrKeyConstructionFailed, err)
	}

	m, err := r.repo.Get(key)
	if err != nil {
		return nil, errors.NewTypedError(ErrTransactionMissing, err)
	}

	return m.(*Transaction), nil
}

// Save saves the transaction to the repository.
func (r *txRepository) Save(tx *Transaction) error {
	key, err := getKey(tx.CID, tx.ID)
	if err != nil {
		return errors.NewTypedError(ErrKeyConstructionFailed, err)
	}

	if r.repo.Exists(key) {
		return r.repo.Update(key, tx)
	}

	return r.repo.Create(key, tx)
}
