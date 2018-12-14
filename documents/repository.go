package documents

import (
	"github.com/centrifuge/go-centrifuge/storage"
)

// Repository defines the required methods for a document repository.
// Can be implemented by any type that stores the documents. Ex: levelDB, sql etc...
type Repository interface {
	// Exists checks if the id, owned by tenantID, exists in DB
	Exists(tenantID, id []byte) bool

	// Get returns the Model associated with ID, owned by tenantID
	Get(tenantID, id []byte) (Model, error)

	// Create creates the model if not present in the DB.
	// should error out if the document exists.
	Create(tenantID, id []byte, model Model) error

	// Update strictly updates the model.
	// Will error out when the model doesn't exist in the DB.
	Update(tenantID, id []byte, model Model) error

	// Register registers the model so that the DB can return the document without knowing the type
	Register(model Model)
}

func NewDBRepository(db storage.Repository) Repository {
	return &repo{db: db}
}

type repo struct {
	db storage.Repository
}

// getKey returns tenantID+id
func (r *repo) getKey(tenantID, id []byte) []byte {
	return append(tenantID, id...)
}

// Register registers the model so that the DB can return the document without knowing the type
func (r *repo) Register(model Model) {
	r.db.Register(model)
}

// Exists checks if the id, owned by tenantID, exists in DB
func (r *repo) Exists(tenantID, id []byte) bool {
	key := r.getKey(tenantID, id)
	return r.db.Exists(key)
}

// Get returns the Model associated with ID, owned by tenantID
func (r *repo) Get(tenantID, id []byte) (Model, error) {
	key := r.getKey(tenantID, id)
	model, err := r.db.Get(key)
	if err != nil {
		return nil, err
	}
	return model.(Model), nil
}

// Create creates the model if not present in the DB.
// should error out if the document exists.
func (r *repo) Create(tenantID, id []byte, model Model) error {
	key := r.getKey(tenantID, id)
	return r.db.Create(key, model)
}

// Update strictly updates the model.
// Will error out when the model doesn't exist in the DB.
func (r *repo) Update(tenantID, id []byte, model Model) error {
	key := r.getKey(tenantID, id)
	return r.db.Update(key, model)
}
