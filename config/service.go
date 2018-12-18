package config

import "github.com/centrifuge/go-centrifuge/protobufs/gen/go/config"

// Service exposes functions over the config objects
type Service interface {
	GetConfig() (*configpb.ConfigData, error)
	GetTenant(identifier []byte) (*configpb.TenantData, error)
	GetAllTenants() ([]*TenantConfig, error)
	CreateConfig(data *configpb.ConfigData) (*configpb.ConfigData, error)
	CreateTenant(data *configpb.TenantData) (*configpb.TenantData, error)
	UpdateConfig(data *configpb.ConfigData) (*configpb.ConfigData, error)
	UpdateTenant(data *configpb.TenantData) (*configpb.TenantData, error)
	DeleteConfig() error
	DeleteTenant(identifier []byte) error
}

type service struct {
	repo Repository
}

func DefaultService(repository Repository) Service {
	return &service{repo: repository}
}

func (s service) GetConfig() (*configpb.ConfigData, error) {
	cfg, err := s.repo.GetConfig()
	if err != nil {
		return nil, err
	}
	return cfg.createProtobuf(), nil
}

func (s service) GetTenant(identifier []byte) (*configpb.TenantData, error) {
	cfg, err := s.repo.GetTenant(identifier)
	if err != nil {
		return nil, err
	}
	return cfg.createProtobuf(), nil
}

func (s service) GetAllTenants() ([]*TenantConfig, error) {
	return s.repo.GetAllTenants()
}

func (s service) CreateConfig(data *configpb.ConfigData) (*configpb.ConfigData, error) {
	nodeConfig := new(NodeConfig)
	nodeConfig.loadFromProtobuf(data)
	err := s.repo.CreateConfig(nodeConfig)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (s service) CreateTenant(data *configpb.TenantData) (*configpb.TenantData, error) {
	tenantConfig := new(TenantConfig)
	tenantConfig.loadFromProtobuf(data)
	err := s.repo.CreateTenant(tenantConfig.IdentityID, tenantConfig)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (s service) UpdateConfig(data *configpb.ConfigData) (*configpb.ConfigData, error) {
	nodeConfig := new(NodeConfig)
	nodeConfig.loadFromProtobuf(data)
	err := s.repo.UpdateConfig(nodeConfig)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (s service) UpdateTenant(data *configpb.TenantData) (*configpb.TenantData, error) {
	tenantConfig := new(TenantConfig)
	tenantConfig.loadFromProtobuf(data)
	err := s.repo.UpdateTenant(tenantConfig.IdentityID, tenantConfig)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (s service) DeleteConfig() error {
	return s.repo.DeleteConfig()
}

func (s service) DeleteTenant(identifier []byte) error {
	return s.repo.DeleteTenant(identifier)
}
