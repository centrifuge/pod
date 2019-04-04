package entityrelationship

import (
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/storage"
)

// repository defines the required methods for the config repository.
type repository interface {
	// Exists checks if the id, owned by accountID, exists in DB
	Exists(accountID, id []byte) bool

	// Get returns the documents.Model associated with ID, owned by accountID
	Get(accountID, id []byte) (*EntityRelationship, error)

	// Create creates the model if not present in the DB.
	// should error out if the document exists.
	Create(accountID, id []byte, model *EntityRelationship) error

	// Update strictly updates the model.
	// Will error out when the model doesn't exist in the DB.
	Update(accountID, id []byte, model EntityRelationship) error

	// Register registers the model so that the DB can return the document without knowing the type
	Register(model *EntityRelationship)

	// Find returns a EntityRelationship based on a entity id and a targetDID
	Find(entityIdentifier []byte,targetDID identity.DID) (*EntityRelationship, error)

	// EntityExists returns true if a entity relationship exists for a given entity identifier and target did
	EntityExists(entityIdentifier []byte,targetDID identity.DID) (bool, error)
}

type repo struct {
	db storage.Repository
}

// newDBRepository creates instance of Config repository
func newDBRepository(db storage.Repository) repository {
	return &repo{db: db}
}

// getKey returns accountID+id
func (r *repo) getKey(accountID, id []byte) []byte {
	return append(accountID, id...)
}

// Register registers the model so that the DB can return the document without knowing the type
func (r *repo) Register(model *EntityRelationship) {
	r.db.Register(model)
}

// Exists checks if the id, owned by accountID, exists in DB
func (r *repo) Exists(accountID, id []byte) bool {
	key := r.getKey(accountID, id)
	return r.db.Exists(key)
}

// Get returns the documents.Model associated with ID, owned by accountID
func (r *repo) Get(accountID, id []byte) (*EntityRelationship, error) {
	key := r.getKey(accountID, id)
	model, err := r.db.Get(key)
	if err != nil {
		return nil, err
	}
	return model.(*EntityRelationship), nil
}

// Create creates the model if not present in the DB.
// should error out if the document exists.
func (r *repo) Create(accountID, id []byte, model *EntityRelationship) error {
	key := r.getKey(accountID, id)
	return r.db.Create(key, model)
}

// Update strictly updates the model.
// Will error out when the model doesn't exist in the DB.
func (r *repo) Update(accountID, id []byte, model EntityRelationship) error {
	key := r.getKey(accountID, id)
	return r.db.Update(key, &model)
}

// Find returns a EntityRelationship based on a entity id and a targetDID
func (r *repo) Find(entityIdentifier []byte,targetDID identity.DID) (*EntityRelationship, error) {
	// todo not implemented
	return nil, nil
}

// EntityExists returns true if a entity relationship exists for a given entity identifier and target did
func (r *repo) EntityExists(entityIdentifier []byte,targetDID identity.DID) (bool, error) {
	// todo not implemented
	return false, nil
}
