package config

import "github.com/centrifuge/go-centrifuge/protobufs/gen/go/config"

// Service exposes functions over the config objects
type Service interface {
	GetConfig() (*configpb.ConfigData, error)
	GetTenant(req *configpb.GetTenantRequest) (*configpb.TenantData, error)
	GetAllTenants() (*configpb.GetAllTenantResponse, error)
	CreateConfig(data *configpb.ConfigData) (*configpb.ConfigData, error)
	CreateTenant(data *configpb.TenantData) (*configpb.TenantData, error)
	UpdateConfig(data *configpb.ConfigData) (*configpb.ConfigData, error)
	UpdateTenant(req *configpb.UpdateTenantRequest) (*configpb.TenantData, error)
	DeleteConfig() error
	DeleteTenant(req *configpb.GetTenantRequest) error
}

type service struct {
	repo Repository
}

func (s service) GetConfig() (*configpb.ConfigData, error) {
	cfg, err := s.repo.GetConfig()
	if err != nil {
		return nil, err
	}
	return cfg.createProtobuf(), nil
}

func (s service) GetTenant(req *configpb.GetTenantRequest) (*configpb.TenantData, error) {
	cfg, err := s.repo.GetTenant([]byte(req.Identifier))
	if err != nil {
		return nil, err
	}
	return cfg.createProtobuf(), nil
}

func convertToAllTenantResponse(cfgs []*TenantConfig) (*configpb.GetAllTenantResponse, error) {
	response := new(configpb.GetAllTenantResponse)
	response.Data = make([]*configpb.TenantData, len(cfgs))
	for _, t := range cfgs {
		response.Data = append(response.Data, t.createProtobuf())
	}
	return response, nil
}

func (s service) GetAllTenants() (*configpb.GetAllTenantResponse, error) {
	cfgs, err := s.repo.GetAllTenants()
	if err != nil {
		return nil, err
	}
	return convertToAllTenantResponse(cfgs)
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

func (s service) UpdateTenant(req *configpb.UpdateTenantRequest) (*configpb.TenantData, error) {
	tenantConfig := new(TenantConfig)
	tenantConfig.loadFromProtobuf(req.Data)
	err := s.repo.UpdateTenant(tenantConfig.IdentityID, tenantConfig)
	if err != nil {
		return nil, err
	}
	return req.Data, nil
}

func (s service) DeleteConfig() error {
	return s.repo.DeleteConfig()
}

func (s service) DeleteTenant(req *configpb.GetTenantRequest) error {
	return s.repo.DeleteTenant([]byte(req.Identifier))
}
