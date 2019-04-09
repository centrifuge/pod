package entityrelationship

import (
	"bytes"

	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/storage"
	"github.com/centrifuge/go-centrifuge/utils"
)

// repository defines the required methods for the config repository.
type repository interface {
	documents.Repository

	// Find returns the latest EntityRelationship based on the document identifier of an Entity and a targetDID
	FindEntityRelationship(entityIdentifier []byte, ownerDID, targetDID identity.DID) (EntityRelationship, error)

	// ListAllRelationships returns a list of all relationships in which a given entity is involved
	ListAllRelationships(entityIdentifier []byte, ownerDID identity.DID) ([]EntityRelationship, error)
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
func (r *repo) FindEntityRelationship(entityIdentifier []byte, ownerDID, targetDID identity.DID) (EntityRelationship, error) {
	relationships, err := r.db.GetAllByPrefix(string(ownerDID[:]))
	if err != nil {
		return EntityRelationship{}, err
	}

	if relationships == nil {
		return EntityRelationship{}, documents.ErrDocumentNotFound
	}

	var versions []EntityRelationship
	for _, r := range relationships {
		e, ok := r.(*EntityRelationship)
		if !ok {
			continue
		}
		if bytes.Equal(e.EntityIdentifier, entityIdentifier) {
			if targetDID.Equal(*e.TargetIdentity) {
				versions = append(versions, *e)
			}
		}
	}
	if len(versions) == 0 {
		return EntityRelationship{}, documents.ErrDocumentNotFound
	}
	if len(versions) > 1 {
		for _, v := range versions {
			if !utils.IsEmptyByteSlice(v.PreviousVersion()) {
				return v, nil
			}
		}
	}
	return versions[0], nil
}

// ListAllRelationships returns a list of all relationships in which a given entity is involved
func (r *repo) ListAllRelationships(entityIdentifier []byte, ownerDID identity.DID) ([]EntityRelationship, error) {
	relationships, err := r.db.GetAllByPrefix(string(ownerDID[:]))
	if err != nil {
		return nil, err
	}

	if relationships == nil {
		return nil, documents.ErrDocumentNotFound
	}

	var all []EntityRelationship
	for _, r := range relationships {
		e, ok := r.(*EntityRelationship)
		if !ok {
			continue
		}
		if bytes.Equal(e.EntityIdentifier, entityIdentifier) {
			all = append(all, *e)
		}
	}
	return all, nil
}
