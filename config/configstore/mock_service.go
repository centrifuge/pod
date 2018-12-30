// +build unit

package configstore

import "github.com/stretchr/testify/mock"

type MockService struct {
	mock.Mock
}

func (MockService) GetConfig() (*NodeConfig, error) {
	return nil, nil
}

func (MockService) GetTenant(identifier []byte) (*TenantConfig, error) {
	panic("implement me")
}

func (MockService) GetAllTenants() ([]*TenantConfig, error) {
	panic("implement me")
}

func (MockService) CreateConfig(data *NodeConfig) (*NodeConfig, error) {
	panic("implement me")
}

func (MockService) CreateTenant(data *TenantConfig) (*TenantConfig, error) {
	panic("implement me")
}

func (MockService) UpdateConfig(data *NodeConfig) (*NodeConfig, error) {
	panic("implement me")
}

func (MockService) UpdateTenant(data *TenantConfig) (*TenantConfig, error) {
	panic("implement me")
}

func (MockService) DeleteConfig() error {
	panic("implement me")
}

func (MockService) DeleteTenant(identifier []byte) error {
	panic("implement me")
}
