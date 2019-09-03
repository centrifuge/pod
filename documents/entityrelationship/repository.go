package entityrelationship

import (
	"bytes"

	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/storage"
	"github.com/ethereum/go-ethereum/common/hexutil"
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

// FindEntityRelationshipIdentifier returns the identifier of an EntityRelationship based on a entity id and a targetDID
func (r *repo) FindEntityRelationshipIdentifier(entityIdentifier []byte, ownerDID, targetDID identity.DID) ([]byte, error) {
	relationships, err := r.db.GetAllByPrefix(documents.DocPrefix + hexutil.Encode(ownerDID[:]))
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
		if bytes.Equal(e.Data.EntityIdentifier, entityIdentifier) && targetDID.Equal(*e.Data.TargetIdentity) {
			return e.ID(), nil
		}
	}
	return nil, documents.ErrDocumentNotFound
}

// ListAllRelationships returns a list of all entity relationship identifiers in which a given entity is involved
func (r *repo) ListAllRelationships(entityIdentifier []byte, ownerDID identity.DID) (map[string][]byte, error) {
	allDocuments, err := r.db.GetAllByPrefix(documents.DocPrefix + hexutil.Encode(ownerDID[:]))
	if err != nil {
		return nil, err
	}

	if allDocuments == nil {
		return map[string][]byte{}, nil
	}

	relationships := make(map[string][]byte)
	for _, r := range allDocuments {
		e, ok := r.(*EntityRelationship)
		if !ok {
			continue
		}
		_, found := relationships[string(e.Document.DocumentIdentifier)]
		if bytes.Equal(e.Data.EntityIdentifier, entityIdentifier) && !found {
			relationships[string(e.Document.DocumentIdentifier)] = e.Document.DocumentIdentifier
		}
	}

	if len(relationships) == 0 {
		return map[string][]byte{}, nil
	}

	return relationships, nil
}
