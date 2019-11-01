package leveldb

import (
	"encoding/json"
	"reflect"
	"sync"

	"github.com/centrifuge/go-centrifuge/storage"

	"github.com/centrifuge/go-centrifuge/errors"
	logging "github.com/ipfs/go-log"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/util"
)

var log = logging.Logger("storage")

// NewLevelDBStorage is a singleton implementation returning the default database as configured.
func NewLevelDBStorage(path string) (*leveldb.DB, error) {
	i, err := leveldb.OpenFile(path, nil)
	if err != nil {
		return nil, err
	}
	return i, nil
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
func NewLevelDBRepository(db *leveldb.DB) storage.Repository {
	return &levelDBRepo{
		db:     db,
		models: make(map[string]reflect.Type),
	}
}

// Register registers the model so that the DB can return the model without knowing the type
func (l *levelDBRepo) Register(model storage.Model) {
	l.mu.Lock()
	defer l.mu.Unlock()
	tp := getTypeIndirect(model.Type())
	l.models[tp.String()] = tp
}

// Exists checks whether the key exists in db
func (l *levelDBRepo) Exists(key []byte) bool {
	res, err := l.db.Has(key, nil)
	if err != nil {
		return false
	}
	return res
}

// getModel returns a new instance of the type mt.
func (l *levelDBRepo) getModel(mt string) (storage.Model, error) {
	tp, ok := l.models[mt]
	if !ok {
		return nil, errors.NewTypedError(storage.ErrModelTypeNotRegistered, errors.New("%s", mt))
	}

	return reflect.New(tp).Interface().(storage.Model), nil
}

func (l *levelDBRepo) parseModel(data []byte) (storage.Model, error) {
	v := new(value)
	err := json.Unmarshal(data, v)
	if err != nil {
		return nil, errors.NewTypedError(storage.ErrModelRepositorySerialisation, errors.New("failed to unmarshal to value: %v", err))
	}

	nm, err := l.getModel(v.Type)
	if err != nil {
		return nil, err
	}

	err = nm.FromJSON(v.Data)
	if err != nil {
		return nil, errors.NewTypedError(storage.ErrModelRepositorySerialisation, errors.New("failed to unmarshal to model: %v", err))
	}

	return nm, nil
}

// Get retrieves model by key, otherwise returns error
func (l *levelDBRepo) Get(key []byte) (storage.Model, error) {
	l.mu.RLock()
	defer l.mu.RUnlock()
	data, err := l.db.Get(key, nil)
	if err != nil {
		return nil, errors.NewTypedError(storage.ErrModelRepositoryNotFound, err)
	}

	return l.parseModel(data)
}

// GetAllByPrefix returns all models which keys match the provided prefix
// If an error is found parsing one of the matched models, logs warning and continues
func (l *levelDBRepo) GetAllByPrefix(prefix string) ([]storage.Model, error) {
	var models []storage.Model
	l.mu.RLock()
	defer l.mu.RUnlock()
	iter := l.db.NewIterator(util.BytesPrefix([]byte(prefix)), nil)
	for iter.Next() {
		data := iter.Value()
		model, err := l.parseModel(data)
		if err != nil {
			log.Warningf("Error parsing model: %v", err)
			continue
		}
		models = append(models, model)
	}
	iter.Release()
	return models, iter.Error()
}

func (l *levelDBRepo) save(key []byte, model storage.Model) error {
	data, err := model.JSON()
	if err != nil {
		return errors.NewTypedError(storage.ErrModelRepositorySerialisation, errors.New("failed to marshall model: %v", err))
	}

	tp := getTypeIndirect(model.Type())
	v := value{
		Type: tp.String(),
		Data: json.RawMessage(data),
	}

	data, err = json.Marshal(v)
	if err != nil {
		return errors.NewTypedError(storage.ErrModelRepositorySerialisation, errors.New("failed to marshall value: %v", err))
	}

	err = l.db.Put(key, data, nil)
	if err != nil {
		return errors.NewTypedError(storage.ErrRepositoryModelSave, errors.New("%v", err))
	}

	return nil
}

// Create creates a model indexed by the key provided
// errors out if key already exists
func (l *levelDBRepo) Create(key []byte, model storage.Model) error {
	if l.Exists(key) {
		return storage.ErrRepositoryModelCreateKeyExists
	}
	return l.save(key, model)
}

// Update updates a model indexed by the key provided
// errors out if key doesn't exists
func (l *levelDBRepo) Update(key []byte, model storage.Model) error {
	if !l.Exists(key) {
		return storage.ErrRepositoryModelUpdateKeyNotFound
	}
	return l.save(key, model)
}

// Delete deletes a model by the key provided
func (l *levelDBRepo) Delete(key []byte) error {
	return l.db.Delete(key, nil)
}

// Close closes the database
func (l *levelDBRepo) Close() error {
	return l.db.Close()
}

// getTypeIndirect returns the type of the model without pointers.
func getTypeIndirect(tp reflect.Type) reflect.Type {
	if tp.Kind() == reflect.Ptr {
		return getTypeIndirect(tp.Elem())
	}

	return tp
}
