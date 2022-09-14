package documents

import (
	"bytes"
	"encoding/json"
	"reflect"
	"time"

	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/storage"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

const (
	// DocPrefix holds the generic prefix of a document in DB
	DocPrefix string = "document_"

	// LatestPrefix is used to index latest version of the document.
	LatestPrefix string = "latest_document_"
)

type latestVersion struct {
	// CurrentVersion is the current latest version of the document
	CurrentVersion []byte `json:"current_version"`

	// NextVersion is the supposed next version of the document.
	// this is stored to check and invalidate the old index
	NextVersion []byte `json:"next_version"`

	// Timestamp is the time when document version is created.
	Timestamp time.Time `json:"timestamp"`
}

// JSON marshals latestVersion to json bytes.
func (l *latestVersion) JSON() ([]byte, error) {
	return json.Marshal(l)
}

// Type returns the type of latestVersion.
func (l *latestVersion) Type() reflect.Type {
	return reflect.TypeOf(l)
}

// FromJSON loads json bytes to latest version
func (l *latestVersion) FromJSON(data []byte) error {
	return json.Unmarshal(data, l)
}

//go:generate mockery --name Repository --structname RepositoryMock --filename repository_mock.go --inpackage

// Repository defines the required methods for a document repository.
// Can be implemented by any type that stores the documents. Ex: levelDB, sql etc...
type Repository interface {
	// Exists checks if the id, owned by accountID, exists in DB
	Exists(accountID, id []byte) bool

	// Get returns the Document associated with ID, owned by accountID
	Get(accountID, id []byte) (Document, error)

	// Create creates the model if not present in the DB.
	// should error out if the document exists.
	Create(accountID, id []byte, model Document) error

	// Update strictly updates the model.
	// Will error out when the model doesn't exist in the DB.
	Update(accountID, id []byte, model Document) error

	// Register registers the model so that the DB can return the document without knowing the type
	Register(model Document)

	// GetLatest returns the latest version of the document.
	GetLatest(accountID, docID []byte) (Document, error)
}

// NewDBRepository creates an instance of the documents Repository
func NewDBRepository(db storage.Repository) Repository {
	db.Register(new(latestVersion))
	return &repo{db: db}
}

type repo struct {
	db storage.Repository
}

// Register registers the model so that the DB can return the document without knowing the type
func (r *repo) Register(model Document) {
	r.db.Register(model)
}

// Exists checks if the id, owned by accountID, exists in DB
func (r *repo) Exists(accountID, id []byte) bool {
	key := getKey(accountID, id)
	return r.db.Exists(key)
}

// Get returns the Document associated with ID, owned by accountID
func (r *repo) Get(accountID, id []byte) (Document, error) {
	key := getKey(accountID, id)
	model, err := r.db.Get(key)
	if err != nil {
		return nil, err
	}
	m, ok := model.(Document)
	if !ok {
		return nil, errors.New("docID %s for account %s is not a model object", hexutil.Encode(id), hexutil.Encode(accountID))
	}
	return m, nil
}

// Create creates the model if not present in the DB.
// should error out if the document exists.
func (r *repo) Create(accountID, id []byte, model Document) error {
	key := getKey(accountID, id)
	if err := r.db.Create(key, model); err != nil {
		return err
	}

	return r.updateLatestIndex(accountID, model)
}

// Update strictly updates the model.
// Will error out when the model doesn't exist in the DB.
func (r *repo) Update(accountID, id []byte, model Document) error {
	key := getKey(accountID, id)
	if err := r.db.Update(key, model); err != nil {
		return err
	}

	return r.updateLatestIndex(accountID, model)
}

// GetLatest returns thee latest version of the document.
func (r *repo) GetLatest(accountID, docID []byte) (Document, error) {
	key := getLatestKey(accountID, docID)
	lv, err := r.getLatestVersion(key)
	if err != nil {
		return nil, err
	}

	return r.Get(accountID, lv.CurrentVersion)
}

func (r *repo) getLatestVersion(key []byte) (*latestVersion, error) {
	val, err := r.db.Get(key)
	if err != nil {
		return nil, err
	}

	lv, ok := val.(*latestVersion)
	if ok {
		return lv, nil
	}

	// delete key val if the type mismatches
	err = r.db.Delete(key)
	if err != nil {
		return nil, err
	}

	return nil, ErrDocumentNotFound
}

// storeLatestIndex stores the latestVersion to db.
// If update is true, it is assumed that index is overwritten
// else, index is created first time.
func (r *repo) storeLatestIndex(key []byte, model Document, update bool) error {
	lv := &latestVersion{
		CurrentVersion: model.CurrentVersion(),
		NextVersion:    model.NextVersion(),
	}

	tm, err := model.Timestamp()
	if err != nil {
		// we will update this actual value when available
		tm = time.Now().UTC()
	}
	lv.Timestamp = tm

	if update {
		return r.db.Update(key, lv)
	}

	return r.db.Create(key, lv)
}

// updateLatestIndex updates the latest version index.
// We check if the latest index is present for a model.
// If not found, create a latest index and return.
// Note: anchor timestamp is not available immediately, so don't error out if the timestamp is empty
// If found, check if the next version matches the current version of the passed model.
// If matches, update the timestamp of anchor and return.
// If not matches, check the model timestamp is greater than stored timestamp.
// If greater update the latestVersion and return
// If not, skip update and return.
func (r *repo) updateLatestIndex(accID []byte, model Document) error {
	if model.GetStatus() != Committed {
		return nil
	}

	key := getLatestKey(accID, model.ID())
	lv, err := r.getLatestVersion(key)
	if err != nil {
		// no index is created yet. create one
		return r.storeLatestIndex(key, model, false)
	}

	if bytes.Equal(lv.NextVersion, model.CurrentVersion()) {
		return r.storeLatestIndex(key, model, true)
	}

	// compare timestamps
	ts, err := model.Timestamp()
	if err != nil {
		ts = time.Now().UTC()
	}

	if lv.Timestamp.Before(ts) {
		// newer version found. so update
		return r.storeLatestIndex(key, model, true)
	}

	// must be an old version.
	return nil
}

// getKey returns document_+accountID+id
func getKey(accountID, id []byte) []byte {
	hexKey := hexutil.Encode(append(accountID, id...))
	return append([]byte(DocPrefix), []byte(hexKey)...)
}

// getLatestKey constructs the key to the latest version of the document.
// Note: DocumentIdentifier needs to be passed here not the versionID.
func getLatestKey(accountID, docID []byte) []byte {
	hexKey := hexutil.Encode(append(accountID, docID...))
	return append([]byte(LatestPrefix), []byte(hexKey)...)
}
