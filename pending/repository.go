package pending

import (
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/storage"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

const (
	// DocPrefix holds the generic prefix of a document in DB
	DocPrefix string = "pending_document_"
)

//go:generate mockery --name Repository --structname RepositoryMock --filename repository_mock.go --inpackage

// Repository defines the required methods for a document repository.
// Can be implemented by any type that stores the documents. Ex: levelDB, sql etc...
type Repository interface {
	// Get returns the Document associated with ID, owned by accountID
	Get(accountID, id []byte) (documents.Document, error)

	// Create creates the model if not present in the DB.
	// should error out if the document exists.
	Create(accountID, id []byte, model documents.Document) error

	// Update strictly updates the model.
	// Will error out when the model doesn't exist in the DB.
	Update(accountID, id []byte, model documents.Document) error

	// Delete deletes the data associated with account and ID.
	Delete(accountID, id []byte) error
}

// NewRepository creates an instance of the pending document Repository
func NewRepository(db storage.Repository) Repository {
	return &repo{db: db}
}

type repo struct {
	db storage.Repository
}

// getKey returns document_+accountID+id
func (r *repo) getKey(accountID, id []byte) []byte {
	hexKey := hexutil.Encode(append(accountID, id...))
	return append([]byte(DocPrefix), []byte(hexKey)...)
}

// Get returns the Document associated with ID, owned by accountID
func (r *repo) Get(accountID, id []byte) (documents.Document, error) {
	key := r.getKey(accountID, id)
	model, err := r.db.Get(key)
	if err != nil {
		return nil, err
	}
	m, ok := model.(documents.Document)
	if !ok {
		return nil, errors.New("docID %s for account %s is not a model object", hexutil.Encode(id), hexutil.Encode(accountID))
	}
	return m, nil
}

// Create creates the model if not present in the DB.
// should error out if the document exists.
func (r *repo) Create(accountID, id []byte, model documents.Document) error {
	key := r.getKey(accountID, id)
	return r.db.Create(key, model)
}

// Update strictly updates the model.
// Will error out when the model doesn't exist in the DB.
func (r *repo) Update(accountID, id []byte, model documents.Document) error {
	key := r.getKey(accountID, id)
	return r.db.Update(key, model)
}

func (r *repo) Delete(accountID, id []byte) error {
	key := r.getKey(accountID, id)
	return r.db.Delete(key)
}
