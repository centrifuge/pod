package entityrelationship

import (
	"bytes"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/storage"
)

// repository defines the required methods for the config repository.
type repository interface {
	documents.Repository

	// Find returns the latest EntityRelationship based on the document identifier of an Entity and a targetDID
	FindEntityRelationship(entityIdentifier []byte, targetDID identity.DID) (*EntityRelationship, error)

	// EntityExists returns true if a entity relationship exists for a given Entity document identifier and target did
	EntityRelationshipExists(entityIdentifier []byte, targetDID identity.DID) (bool, error)

	// ListAllRelationships returns a list of all relationships in which a given entity is involved
	ListAllRelationships(entityIdentifier []byte) ([]EntityRelationship, error)
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
func (r *repo) FindEntityRelationship(entityIdentifier []byte, targetDID identity.DID) (*EntityRelationship, error) {
	relationships, err := r.db.GetAllByPrefix("")
	if err != nil {
		return nil, err
	}

	if relationships == nil {
		return nil, documents.ErrDocumentNotFound
	}

	for _, r := range relationships {
		if r.(*EntityRelationship).TargetIdentity == &targetDID {
			if r.(*EntityRelationship).Document.AccessTokens != nil {
				if bytes.Equal(r.(*EntityRelationship).Document.AccessTokens[0].DocumentIdentifier, entityIdentifier) {
					return r.(*EntityRelationship), nil
				}
			}
		}
	}
	return nil, documents.ErrDocumentNotFound
}

// EntityExists returns true if a entity relationship exists for a given entity identifier and target did
func (r *repo) EntityRelationshipExists(entityIdentifier []byte, targetDID identity.DID) (bool, error) {
	relationships, err := r.db.GetAllByPrefix(prefix)
	if err != nil {
		return false, err
	}

	if relationships == nil {
		return false, documents.ErrDocumentNotFound
	}

	for _, r := range relationships {
		if r.(*EntityRelationship).TargetIdentity == &targetDID {
			if r.(*EntityRelationship).Document.AccessTokens != nil {
				if bytes.Equal(r.(*EntityRelationship).Document.AccessTokens[0].DocumentIdentifier, entityIdentifier) {
					return true, nil
				}
			}
		}
	}
	return false, documents.ErrDocumentNotFound
}

// ListAllRelationships returns a list of all relationships in which a given entity is involved
func (r *repo) ListAllRelationships(entityIdentifier []byte) ([]EntityRelationship, error) {
	relationships, err := r.db.GetAllByPrefix(prefix)
	if err != nil {
		return nil, err
	}

	if relationships == nil {
		return nil, documents.ErrDocumentNotFound
	}

	var er []EntityRelationship
	for _, r := range relationships {
		if r.(*EntityRelationship).Document.AccessTokens != nil {
			if bytes.Equal(r.(*EntityRelationship).Document.AccessTokens[0].DocumentIdentifier, entityIdentifier) {
				er = append(er, *r.(*EntityRelationship))
			}
		}
	}
	return er, nil
}
