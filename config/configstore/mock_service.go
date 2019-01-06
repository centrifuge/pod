// +build unit

package configstore

import "github.com/stretchr/testify/mock"

type MockService struct {
	mock.Mock
}

func (m MockService) GetConfig() (*NodeConfig, error) {
	args := m.Called()
	return args.Get(0).(*NodeConfig), args.Error(1)
}

func (m MockService) GetTenant(identifier []byte) (*TenantConfig, error) {
	args := m.Called(identifier)
	return args.Get(0).(*TenantConfig), args.Error(0)
}

func (m MockService) GetAllTenants() ([]*TenantConfig, error) {
	args := m.Called()
	v, _ := args.Get(0).([]*TenantConfig)
	return v, nil
}

func (m MockService) CreateConfig(data *NodeConfig) (*NodeConfig, error) {
	args := m.Called(data)
	return args.Get(0).(*NodeConfig), args.Error(0)
}

func (m MockService) CreateTenant(data *TenantConfig) (*TenantConfig, error) {
	args := m.Called(data)
	return args.Get(0).(*TenantConfig), args.Error(0)
}

func (m MockService) UpdateConfig(data *NodeConfig) (*NodeConfig, error) {
	args := m.Called()
	return args.Get(0).(*NodeConfig), args.Error(0)
}

func (m MockService) UpdateTenant(data *TenantConfig) (*TenantConfig, error) {
	args := m.Called(data)
	return args.Get(0).(*TenantConfig), args.Error(0)
}

func (m MockService) DeleteConfig() error {
	args := m.Called()
	return args.Error(0)
}

func (m MockService) DeleteTenant(identifier []byte) error {
	args := m.Called(identifier)
	return args.Error(0)
}
