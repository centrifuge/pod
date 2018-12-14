package storage

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sync"

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
func NewLevelDBRepository(db *leveldb.DB) Repository {
	return &levelDBRepo{
		db:     db,
		models: make(map[string]reflect.Type),
	}
}

// Register registers the model so that the DB can return the model without knowing the type
func (l *levelDBRepo) Register(model Model) {
	l.mu.Lock()
	defer l.mu.Unlock()
	tp := getTypeIndirect(model.Type())
	l.models[tp.String()] = tp
}

func (l *levelDBRepo) Exists(key []byte) bool {
	res, err := l.db.Has(key, nil)
	if err != nil {
		return false
	}
	return res
}

// getModel returns a new instance of the type mt.
func (l *levelDBRepo) getModel(mt string) (Model, error) {
	tp, ok := l.models[mt]
	if !ok {
		return nil, fmt.Errorf("type %s not registered", mt)
	}

	return reflect.New(tp).Interface().(Model), nil
}

func (l *levelDBRepo) parseModel(data []byte) (Model, error) {
	v := new(value)
	err := json.Unmarshal(data, v)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal value: %v", err)
	}

	nm, err := l.getModel(v.Type)
	if err != nil {
		return nil, fmt.Errorf("failed to get model type: %v", err)
	}

	err = nm.FromJSON([]byte(v.Data))
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal to model: %v", err)
	}

	return nm, nil
}

func (l *levelDBRepo) Get(key []byte) (Model, error) {
	l.mu.RLock()
	defer l.mu.RUnlock()
	data, err := l.db.Get(key, nil)
	if err != nil {
		return nil, fmt.Errorf("model missing: %v", err)
	}

	return l.parseModel(data)
}

func (l *levelDBRepo) GetAllByPrefix(prefix string) ([]Model, error) {
	var models []Model
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

func (l *levelDBRepo) save(key []byte, model Model) error {
	data, err := model.JSON()
	if err != nil {
		return errors.NewTypedError(ErrModelRepositorySerialisation, errors.New("failed to marshall model: %v", err))
	}

	tp := getTypeIndirect(model.Type())
	v := value{
		Type: tp.String(),
		Data: json.RawMessage(data),
	}

	data, err = json.Marshal(v)
	if err != nil {
		return errors.NewTypedError(ErrModelRepositorySerialisation, errors.New("failed to marshall value: %v", err))
	}

	err = l.db.Put(key, data, nil)
	if err != nil {
		return errors.NewTypedError(ErrRepositoryModelSave, errors.New("%v", err))
	}

	return nil
}

func (l *levelDBRepo) Create(key []byte, model Model) error {
	if l.Exists(key) {
		return ErrRepositoryModelCreateKeyExists
	}
	return l.save(key, model)
}

func (l *levelDBRepo) Update(key []byte, model Model) error {
	if !l.Exists(key) {
		return ErrRepositoryModelUpdateKeyNotFound
	}
	return l.save(key, model)
}

func (l *levelDBRepo) Delete(key []byte) error {
	return l.db.Delete(key, nil)
}

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
