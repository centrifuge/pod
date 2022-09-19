package entityrelationship

import (
	"bytes"

	"github.com/centrifuge/go-centrifuge/errors"

	"github.com/centrifuge/go-substrate-rpc-client/v4/types"

	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/storage"
)

//go:generate mockery --name repository --structname repositoryMock --filename repository_mock.go --inpackage

// repository defines the required methods for the config repository.
type repository interface {
	documents.Repository

	// FindEntityRelationshipIdentifier returns the latest EntityRelationship based on the document identifier of an Entity and a targetDID
	FindEntityRelationshipIdentifier(entityIdentifier []byte, ownerAccountID, targetAccountID *types.AccountID) ([]byte, error)

	// ListAllRelationships returns a list of all relationships in which a given entity is involved
	ListAllRelationships(entityIdentifier []byte, ownerAccountID *types.AccountID) (map[string][]byte, error)
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

// FindEntityRelationshipIdentifier returns the identifier of an EntityRelationship based on an entity id and a targetDID
func (r *repo) FindEntityRelationshipIdentifier(entityIdentifier []byte, ownerAccountID, targetAccountID *types.AccountID) ([]byte, error) {
	relationships, err := r.db.GetAllByPrefix(documents.DocPrefix + ownerAccountID.ToHexString())
	if err != nil {
		return nil, errors.NewTypedError(ErrDocumentsStorageRetrieval, err)
	}

	if relationships == nil {
		return nil, documents.ErrDocumentNotFound
	}

	for _, r := range relationships {
		e, ok := r.(*EntityRelationship)
		if !ok {
			continue
		}
		if bytes.Equal(e.Data.EntityIdentifier, entityIdentifier) && targetAccountID.Equal(e.Data.TargetIdentity) {
			return e.ID(), nil
		}
	}
	return nil, documents.ErrDocumentNotFound
}

// ListAllRelationships returns a list of all entity relationship identifiers in which a given entity is involved
func (r *repo) ListAllRelationships(entityIdentifier []byte, ownerAccountID *types.AccountID) (map[string][]byte, error) {
	allDocuments, err := r.db.GetAllByPrefix(documents.DocPrefix + ownerAccountID.ToHexString())
	if err != nil {
		return nil, errors.NewTypedError(ErrDocumentsStorageRetrieval, err)
	}

	if allDocuments == nil {
		return nil, nil
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
		return nil, nil
	}

	return relationships, nil
}
