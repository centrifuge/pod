package documents

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sync"

	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/syndtr/goleveldb/leveldb"
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

// levelDBRepo implements Repository using LevelDB as storage layer
type levelDBRepo struct {
	db     *leveldb.DB
	models map[string]reflect.Type
	mu     sync.RWMutex // to protect the models
}

// value is an internal representation of how levelDb stores the model.
type value struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

// NewLevelDBRepository returns levelDb implementation of Repository
func NewLevelDBRepository(db *leveldb.DB) Repository {
	return &levelDBRepo{
		db:     db,
		models: make(map[string]reflect.Type),
	}
}

// getKey returns tenantID+id
func getKey(tenantID, id []byte) []byte {
	return append(tenantID, id...)
}

// Exists returns true if the id, owned by tenantID, exists.
func (l *levelDBRepo) Exists(tenantID, id []byte) bool {
	key := getKey(tenantID, id)
	res, err := l.db.Has(key, nil)
	// TODO check this
	if err != nil {
		return false
	}

	return res
}

// getModel returns a new instance of the type mt.
func (l *levelDBRepo) getModel(mt string) (Model, error) {
	tp, ok := l.models[mt]
	if !ok {
		return nil, errors.NewTypedError(ErrDocumentRepositoryModelNotRegistered, fmt.Errorf("type %s not registered", mt))
	}

	return reflect.New(tp).Interface().(Model), nil
}

// Get returns the model associated with ID, owned by tenantID.
func (l *levelDBRepo) Get(tenantID, id []byte) (Model, error) {
	key := getKey(tenantID, id)
	data, err := l.db.Get(key, nil)
	if err != nil {
		return nil, errors.NewTypedError(ErrDocumentRepositoryModelNotFound, fmt.Errorf("document missing: %v", err))
	}

	v := new(value)
	err = json.Unmarshal(data, v)
	if err != nil {
		return nil, errors.NewTypedError(ErrDocumentRepositorySerialisation, fmt.Errorf("failed to unmarshal value: %v", err))
	}

	l.mu.RLock()
	defer l.mu.RUnlock()
	nm, err := l.getModel(v.Type)
	if err != nil {
		return nil, err
	}

	err = nm.FromJSON([]byte(v.Data))
	if err != nil {
		return nil, errors.NewTypedError(ErrDocumentRepositorySerialisation, fmt.Errorf("failed to unmarshal to model: %v", err))
	}

	return nm, nil
}

// getTypeIndirect returns the type of the model without pointers.
func getTypeIndirect(tp reflect.Type) reflect.Type {
	if tp.Kind() == reflect.Ptr {
		return getTypeIndirect(tp.Elem())
	}

	return tp
}

// save stores the model.
func (l *levelDBRepo) save(tenantID, id []byte, model Model) error {
	data, err := model.JSON()
	if err != nil {
		return errors.NewTypedError(ErrDocumentRepositorySerialisation, fmt.Errorf("failed to marshall model: %v", err))
	}

	tp := getTypeIndirect(model.Type())
	v := value{
		Type: tp.String(),
		Data: json.RawMessage(data),
	}

	data, err = json.Marshal(v)
	if err != nil {
		return errors.NewTypedError(ErrDocumentRepositorySerialisation, fmt.Errorf("failed to marshall value: %v", err))
	}

	key := getKey(tenantID, id)
	err = l.db.Put(key, data, nil)
	if err != nil {
		return errors.NewTypedError(ErrDocumentRepositoryModelSave, fmt.Errorf("%v", err))
	}

	return nil
}

// Create stores the model to the DB.
// Errors out if the model already exists.
func (l *levelDBRepo) Create(tenantID, id []byte, model Model) error {
	if l.Exists(tenantID, id) {
		return ErrDocumentRepositoryModelAllReadyExists
	}

	return l.save(tenantID, id, model)
}

// Update overwrites the value at tenantID+id.
// Errors out if model doesn't exist
func (l *levelDBRepo) Update(tenantID, id []byte, model Model) error {
	if !l.Exists(tenantID, id) {
		return ErrDocumentRepositoryModelDoesntExist
	}

	return l.save(tenantID, id, model)
}

// Register registers the model for type less operations.
// Same type names will be overwritten.
func (l *levelDBRepo) Register(model Model) {
	l.mu.Lock()
	defer l.mu.Unlock()
	tp := getTypeIndirect(model.Type())
	l.models[tp.String()] = tp
}
