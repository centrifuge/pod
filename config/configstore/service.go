package configstore

import (
	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/errors"
)

type service struct {
	repo Repository
}

// DefaultService returns an implementation of the config.Service
func DefaultService(repository Repository) config.Service {
	return &service{repo: repository}
}

func (s service) GetConfig() (config.Configuration, error) {
	return s.repo.GetConfig()
}

func (s service) GetTenant(identifier []byte) (config.TenantConfiguration, error) {
	return s.repo.GetTenant(identifier)
}

func (s service) GetAllTenants() ([]config.TenantConfiguration, error) {
	return s.repo.GetAllTenants()
}

func (s service) CreateConfig(data config.Configuration) (config.Configuration, error) {
	return data, s.repo.CreateConfig(data)
}

func (s service) CreateTenant(data config.TenantConfiguration) (config.TenantConfiguration, error) {
	id, err := data.GetIdentityID()
	if err != nil {
		return nil, err
	}
	return data, s.repo.CreateTenant(id, data)
}

func (s service) UpdateConfig(data config.Configuration) (config.Configuration, error) {
	return data, s.repo.UpdateConfig(data)
}

func (s service) UpdateTenant(data config.TenantConfiguration) (config.TenantConfiguration, error) {
	id, err := data.GetIdentityID()
	if err != nil {
		return nil, err
	}
	return data, s.repo.UpdateTenant(id, data)
}

func (s service) DeleteConfig() error {
	return s.repo.DeleteConfig()
}

func (s service) DeleteTenant(identifier []byte) error {
	return s.repo.DeleteTenant(identifier)
}

// RetrieveConfig retrieves system config giving priority to db stored config
func RetrieveConfig(dbOnly bool, ctx map[string]interface{}) (config.Configuration, error) {
	var cfg config.Configuration
	var err error
	if cfgService, ok := ctx[BootstrappedConfigStorage].(config.Service); ok {
		// may be we need a way to detect a corrupted db here
		cfg, err = cfgService.GetConfig()
		if err != nil {
			apiLog.Warningf("could not load config from db: %v", err)
		}
		return cfg, nil
	}

	// we have to allow loading from file in case this is coming from create config cmd where we don't add configs to db
	if _, ok := ctx[bootstrap.BootstrappedConfig]; ok && cfg == nil && !dbOnly {
		cfg = ctx[bootstrap.BootstrappedConfig].(config.Configuration)
	} else {
		return nil, errors.NewTypedError(ErrConfigRetrieve, err)
	}
	return cfg, nil
}
