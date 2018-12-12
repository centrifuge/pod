package config

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sync"

	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/util"
)

const (
	ConfigPrefix string = "config"
	TenantPrefix string = "tenant-"
)

// Repository defines the required methods for the config repository.
type Repository interface {
	// Get returns the tenant config Model associated with tenant ID
	GetTenant(id []byte) (Model, error)

	// GetConfig returns the node config model
	GetConfig() (Model, error)

	// GetAllTenants returns a list of all tenant models in the config DB
	GetAllTenants() ([]Model, error)

	// Create creates the tenant config model if not present in the DB.
	// should error out if the config exists.
	CreateTenant(id []byte, model Model) error

	// Create creates the node config model if not present in the DB.
	// should error out if the config exists.
	CreateConfig(model Model) error

	// Update strictly updates the tenant config model.
	// Will error out when the config model doesn't exist in the DB.
	UpdateTenant(id []byte, model Model) error

	// Update strictly updates the node config model.
	// Will error out when the config model doesn't exist in the DB.
	UpdateConfig(model Model) error

	// Delete deletes tenant config
	// Will not error out when config model doesn't exists in DB
	DeleteTenant(id []byte) error

	// Delete deletes node config
	// Will not error out when config model doesn't exists in DB
	DeleteConfig() error

	// Register registers the model so that the DB can return the config without knowing the type
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

func (l *levelDBRepo) getTenantKey(id []byte) []byte {
	return append([]byte(TenantPrefix), id...)
}

func (l *levelDBRepo) getConfigKey() []byte {
	return []byte(ConfigPrefix)
}

// getModel returns a new instance of the type mt.
func (l *levelDBRepo) getModel(mt string) (Model, error) {
	tp, ok := l.models[mt]
	if !ok {
		return nil, fmt.Errorf("type %s not registered", mt)
	}

	return reflect.New(tp).Interface().(Model), nil
}

func (l *levelDBRepo) GetTenant(id []byte) (Model, error) {
	key := l.getTenantKey(id)
	return l.get(key)
}

func (l *levelDBRepo) GetConfig() (Model, error) {
	key := l.getConfigKey()
	return l.get(key)
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

// Get returns the model associated with ID
func (l *levelDBRepo) get(id []byte) (Model, error) {
	l.mu.RLock()
	defer l.mu.RUnlock()
	data, err := l.db.Get(id, nil)
	if err != nil {
		return nil, fmt.Errorf("config missing: %v", err)
	}

	return l.parseModel(data)
}

// GetAllTenants iterates over all tenant entries in DB and returns a list of Models
// If an error occur reading a tenant, throws a warning and continue
func (l *levelDBRepo) GetAllTenants() ([]Model, error) {
	var models []Model
	l.mu.RLock()
	defer l.mu.RUnlock()
	iter := l.db.NewIterator(util.BytesPrefix([]byte(TenantPrefix)), nil)
	for iter.Next() {
		data := iter.Value()
		model, err := l.parseModel(data)
		if err != nil {
			log.Warningf("Error parsing tenant: %v", err)
			continue
		}
		models = append(models, model)
	}
	iter.Release()
	return models, iter.Error()
}

// save stores the model.
func (l *levelDBRepo) save(id []byte, model Model) error {
	data, err := model.JSON()
	if err != nil {
		return fmt.Errorf("failed to marshall model: %v", err)
	}

	tp := getTypeIndirect(model.Type())
	v := value{
		Type: tp.String(),
		Data: json.RawMessage(data),
	}

	data, err = json.Marshal(v)
	if err != nil {
		return fmt.Errorf("failed to marshall value: %v", err)
	}

	err = l.db.Put(id, data, nil)
	if err != nil {
		return fmt.Errorf("failed to save model to DB: %v", err)
	}

	return nil
}

// Exists returns true if the id exists.
func (l *levelDBRepo) exists(id []byte) bool {
	res, err := l.db.Has(id, nil)
	// TODO check this
	if err != nil {
		return false
	}

	return res
}

// Create creates the tenant config model if not present in the DB.
// should error out if the config exists.
func (l *levelDBRepo) CreateTenant(id []byte, model Model) error {
	key := l.getTenantKey(id)
	return l.create(key, model)
}

// Create creates the node config model if not present in the DB.
// should error out if the config exists.
func (l *levelDBRepo) CreateConfig(model Model) error {
	key := l.getConfigKey()
	return l.create(key, model)
}

// Create stores the model to the DB.
// Errors out if the model already exists.
func (l *levelDBRepo) create(id []byte, model Model) error {
	if l.exists(id) {
		return fmt.Errorf("model already exists")
	}

	return l.save(id, model)
}

// Update strictly updates the tenant config model.
// Will error out when the config model doesn't exist in the DB.
func (l *levelDBRepo) UpdateTenant(id []byte, model Model) error {
	key := l.getTenantKey(id)
	return l.update(key, model)
}

// Update strictly updates the node config model.
// Will error out when the config model doesn't exist in the DB.
func (l *levelDBRepo) UpdateConfig(model Model) error {
	key := l.getConfigKey()
	return l.update(key, model)
}

// Update overwrites the value at tenantID+id.
// Errors out if model doesn't exist
func (l *levelDBRepo) update(id []byte, model Model) error {
	if !l.exists(id) {
		return fmt.Errorf("model doesn't exist")
	}

	return l.save(id, model)
}

// Delete deletes tenant config
// Will not error out when config model doesn't exists in DB
func (l *levelDBRepo) DeleteTenant(id []byte) error {
	key := l.getTenantKey(id)
	return l.db.Delete(key, nil)
}

func (l *levelDBRepo) DeleteConfig() error {
	key := l.getConfigKey()
	return l.db.Delete(key, nil)
}

// Register registers the model for type less operations.
// Same type names will be overwritten.
func (l *levelDBRepo) Register(model Model) {
	l.mu.Lock()
	defer l.mu.Unlock()
	tp := getTypeIndirect(model.Type())
	l.models[tp.String()] = tp
}

// getTypeIndirect returns the type of the model without pointers.
func getTypeIndirect(tp reflect.Type) reflect.Type {
	if tp.Kind() == reflect.Ptr {
		return getTypeIndirect(tp.Elem())
	}

	return tp
}
