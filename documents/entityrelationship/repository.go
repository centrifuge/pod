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
	FindEntityRelationshipIdentifier(entityIdentifier []byte, ownerDID, targetDID identity.DID) ([]byte, error)

	// ListAllRelationships returns a list of all relationships in which a given entity is involved
	ListAllRelationships(entityIdentifier []byte, ownerDID identity.DID) (map[string][]byte, error)
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

// Find returns the latest (second if revoked) version of an EntityRelationship based on a entity id and a targetDID
// Note that we assume a case of maximum two versions of an EntityRelationship document
func (r *repo) FindEntityRelationshipIdentifier(entityIdentifier []byte, ownerDID, targetDID identity.DID) ([]byte, error) {
	relationships, err := r.db.GetAllByPrefix(string(ownerDID[:]))
	if err != nil {
		return nil, err
	}

	if relationships == nil {
		return nil, documents.ErrDocumentNotFound
	}

	for _, r := range relationships {
		e, ok := r.(*EntityRelationship)
		if !ok {
			continue
		}
		if bytes.Equal(e.EntityIdentifier, entityIdentifier) {
			if targetDID.Equal(*e.TargetIdentity) {
				return e.Document.DocumentIdentifier, nil
			}
		}
	}
	return nil, documents.ErrDocumentNotFound
}

// ListAllRelationships returns a list of all entity relationship identifiers in which a given entity is involved
func (r *repo) ListAllRelationships(entityIdentifier []byte, ownerDID identity.DID) (map[string][]byte, error) {
	relationships, err := r.db.GetAllByPrefix(string(ownerDID[:]))
	if err != nil {
		return nil, err
	}

	if relationships == nil {
		return nil, documents.ErrDocumentNotFound
	}

	all := make(map[string][]byte)
	for _, r := range relationships {
		e, ok := r.(*EntityRelationship)
		if !ok {
			continue
		}
		_, found := all[string(e.Document.DocumentIdentifier)]
		if bytes.Equal(e.EntityIdentifier, entityIdentifier) && !found {
			all[string(e.Document.DocumentIdentifier)] = e.Document.DocumentIdentifier
		}
	}

	if len(all) == 0 {
		return nil, documents.ErrDocumentNotFound
	}

	return all, nil
}
