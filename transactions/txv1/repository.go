package txv1

import (
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/storage"
	"github.com/centrifuge/go-centrifuge/transactions"
)

// txRepository implements Repository.
type txRepository struct {
	repo storage.Repository
}

// NewRepository registers the the transaction model and returns the an implementation
// of the Repository.
func NewRepository(repo storage.Repository) transactions.Repository {
	repo.Register(&transactions.Transaction{})
	return &txRepository{repo: repo}
}

// getKey appends identity with id.
// With identity coming at first, we can even fetch transactions belonging to specific identity through prefix.
func getKey(cid identity.DID, id transactions.TxID) ([]byte, error) {
	if transactions.TxIDEqual(transactions.NilTxID(), id) {
		return nil, errors.New("transaction ID is not valid")
	}

	return append(cid[:], id.Bytes()...), nil
}

// Get returns the transaction associated with identity and id.
func (r *txRepository) Get(cid identity.DID, id transactions.TxID) (*transactions.Transaction, error) {
	key, err := getKey(cid, id)
	if err != nil {
		return nil, errors.NewTypedError(transactions.ErrKeyConstructionFailed, err)
	}

	m, err := r.repo.Get(key)
	if err != nil {
		return nil, errors.NewTypedError(transactions.ErrTransactionMissing, err)
	}

	return m.(*transactions.Transaction), nil
}

// Save saves the transaction to the repository.
func (r *txRepository) Save(tx *transactions.Transaction) error {
	key, err := getKey(tx.DID, tx.ID)
	if err != nil {
		return errors.NewTypedError(transactions.ErrKeyConstructionFailed, err)
	}

	if r.repo.Exists(key) {
		return r.repo.Update(key, tx)
	}

	return r.repo.Create(key, tx)
}
