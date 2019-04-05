package entityrelationship

import (
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/storage"
)

// repository defines the required methods for the config repository.
type repository interface {
	documents.Repository

	// Find returns a EntityRelationship based on a entity id and a targetDID
	Find(entityIdentifier []byte, targetDID identity.DID) (*EntityRelationship, error)

	// EntityExists returns true if a entity relationship exists for a given entity identifier and target did
	EntityExists(entityIdentifier []byte, targetDID identity.DID) (bool, error)
}

type repo struct {
	documents.Repository
	db storage.Repository
}

// newDBRepository creates instance of Config repository
func newDBRepository(db storage.Repository, docRepo documents.Repository) repository {
	r := &repo{db: db}
	r.Repository = docRepo
	return r
}

// Find returns a EntityRelationship based on a entity id and a targetDID
func (r *repo) Find(entityIdentifier []byte, targetDID identity.DID) (*EntityRelationship, error) {
	// todo not implemented
	return nil, nil
}

// EntityExists returns true if a entity relationship exists for a given entity identifier and target did
func (r *repo) EntityExists(entityIdentifier []byte, targetDID identity.DID) (bool, error) {
	// todo not implemented
	return false, nil
}
