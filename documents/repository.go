package documents

import (
	"github.com/centrifuge/go-centrifuge/storage"
)

// Repository defines the required methods for a Document repository.
// Can be implemented by any type that stores the documents. Ex: levelDB, sql etc...
type Repository interface {
	// Exists checks if the id, owned by accountID, exists in DB
	Exists(accountID, id []byte) bool

	// Get returns the Model associated with ID, owned by accountID
	Get(accountID, id []byte) (Model, error)

	// Create creates the model if not present in the DB.
	// should error out if the Document exists.
	Create(accountID, id []byte, model Model) error

	// Update strictly updates the model.
	// Will error out when the model doesn't exist in the DB.
	Update(accountID, id []byte, model Model) error

	// Register registers the model so that the DB can return the Document without knowing the type
	Register(model Model)
}

// NewDBRepository creates an instance of the documents Repository
func NewDBRepository(db storage.Repository) Repository {
	return &repo{db: db}
}

type repo struct {
	db storage.Repository
}

// getKey returns accountID+id
func (r *repo) getKey(accountID, id []byte) []byte {
	return append(accountID, id...)
}

// Register registers the model so that the DB can return the Document without knowing the type
func (r *repo) Register(model Model) {
	r.db.Register(model)
}

// Exists checks if the id, owned by accountID, exists in DB
func (r *repo) Exists(accountID, id []byte) bool {
	key := r.getKey(accountID, id)
	return r.db.Exists(key)
}

// Get returns the Model associated with ID, owned by accountID
func (r *repo) Get(accountID, id []byte) (Model, error) {
	key := r.getKey(accountID, id)
	model, err := r.db.Get(key)
	if err != nil {
		return nil, err
	}
	return model.(Model), nil
}

// Create creates the model if not present in the DB.
// should error out if the Document exists.
func (r *repo) Create(accountID, id []byte, model Model) error {
	key := r.getKey(accountID, id)
	return r.db.Create(key, model)
}

// Update strictly updates the model.
// Will error out when the model doesn't exist in the DB.
func (r *repo) Update(accountID, id []byte, model Model) error {
	key := r.getKey(accountID, id)
	return r.db.Update(key, model)
}
