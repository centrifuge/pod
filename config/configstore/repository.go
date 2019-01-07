package configstore

import (
	"github.com/centrifuge/go-centrifuge/storage"
)

const (
	configPrefix string = "config"
	tenantPrefix string = "tenant-"
)

// Repository defines the required methods for the config repository.
type Repository interface {
	// RegisterTenant registers tenant config in DB
	RegisterTenant(config *TenantConfig)

	// RegisterConfig registers node config in DB
	RegisterConfig(config *NodeConfig)

	// GetTenant returns the tenant config Model associated with tenant ID
	GetTenant(id []byte) (*TenantConfig, error)

	// GetConfig returns the node config model
	GetConfig() (*NodeConfig, error)

	// GetAllTenants returns a list of all tenant models in the config DB
	GetAllTenants() ([]*TenantConfig, error)

	// Create creates the tenant config model if not present in the DB.
	// should error out if the config exists.
	CreateTenant(id []byte, tenant *TenantConfig) error

	// Create creates the node config model if not present in the DB.
	// should error out if the config exists.
	CreateConfig(nodeConfig *NodeConfig) error

	// Update strictly updates the tenant config model.
	// Will error out when the config model doesn't exist in the DB.
	UpdateTenant(id []byte, tenant *TenantConfig) error

	// Update strictly updates the node config model.
	// Will error out when the config model doesn't exist in the DB.
	UpdateConfig(nodeConfig *NodeConfig) error

	// Delete deletes tenant config
	// Will not error out when config model doesn't exists in DB
	DeleteTenant(id []byte) error

	// Delete deletes node config
	// Will not error out when config model doesn't exists in DB
	DeleteConfig() error
}

type repo struct {
	db storage.Repository
}

func getTenantKey(id []byte) []byte {
	return append([]byte(tenantPrefix), id...)
}

func getConfigKey() []byte {
	return []byte(configPrefix)
}

// NewDBRepository creates instance of Config Repository
func NewDBRepository(db storage.Repository) Repository {
	return &repo{db: db}
}

// RegisterTenant registers tenant config in DB
func (r *repo) RegisterTenant(config *TenantConfig) {
	r.db.Register(config)
}

// RegisterConfig registers node config in DB
func (r *repo) RegisterConfig(config *NodeConfig) {
	r.db.Register(config)
}

// GetTenant returns the tenant config Model associated with tenant ID
func (r *repo) GetTenant(id []byte) (*TenantConfig, error) {
	key := getTenantKey(id)
	model, err := r.db.Get(key)
	if err != nil {
		return nil, err
	}
	return model.(*TenantConfig), nil
}

// GetConfig returns the node config model
func (r *repo) GetConfig() (*NodeConfig, error) {
	key := getConfigKey()
	model, err := r.db.Get(key)
	if err != nil {
		return nil, err
	}
	return model.(*NodeConfig), nil
}

// GetAllTenants iterates over all tenant entries in DB and returns a list of Models
// If an error occur reading a tenant, throws a warning and continue
func (r *repo) GetAllTenants() ([]*TenantConfig, error) {
	var tenantConfigs []*TenantConfig
	models, err := r.db.GetAllByPrefix(tenantPrefix)
	if err != nil {
		return nil, err
	}
	for _, tc := range models {
		tenantConfigs = append(tenantConfigs, tc.(*TenantConfig))
	}
	return tenantConfigs, nil
}

// Create creates the tenant config model if not present in the DB.
// should error out if the config exists.
func (r *repo) CreateTenant(id []byte, tenant *TenantConfig) error {
	key := getTenantKey(id)
	return r.db.Create(key, tenant)
}

// Create creates the node config model if not present in the DB.
// should error out if the config exists.
func (r *repo) CreateConfig(nodeConfig *NodeConfig) error {
	key := getConfigKey()
	return r.db.Create(key, nodeConfig)
}

// Update strictly updates the tenant config model.
// Will error out when the config model doesn't exist in the DB.
func (r *repo) UpdateTenant(id []byte, tenant *TenantConfig) error {
	key := getTenantKey(id)
	return r.db.Update(key, tenant)
}

// Update strictly updates the node config model.
// Will error out when the config model doesn't exist in the DB.
func (r *repo) UpdateConfig(nodeConfig *NodeConfig) error {
	key := getConfigKey()
	return r.db.Update(key, nodeConfig)
}

// Delete deletes tenant config
// Will not error out when config model doesn't exists in DB
func (r *repo) DeleteTenant(id []byte) error {
	key := getTenantKey(id)
	return r.db.Delete(key)
}

// Delete deletes node config
// Will not error out when config model doesn't exists in DB
func (r *repo) DeleteConfig() error {
	key := getConfigKey()
	return r.db.Delete(key)
}
