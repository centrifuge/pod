// +build unit

package configstore

import (
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/stretchr/testify/mock"
)

type MockService struct {
	mock.Mock
}

func (m MockService) GenerateTenant() (config.TenantConfiguration, error) {
	args := m.Called()
	return args.Get(0).(config.TenantConfiguration), args.Error(1)
}

func (m MockService) GetConfig() (config.Configuration, error) {
	args := m.Called()
	return args.Get(0).(*NodeConfig), args.Error(1)
}

func (m MockService) GetTenant(identifier []byte) (config.TenantConfiguration, error) {
	args := m.Called(identifier)
	return args.Get(0).(*TenantConfig), args.Error(0)
}

func (m MockService) GetAllTenants() ([]config.TenantConfiguration, error) {
	args := m.Called()
	v, _ := args.Get(0).([]config.TenantConfiguration)
	return v, nil
}

func (m MockService) CreateConfig(data config.Configuration) (config.Configuration, error) {
	args := m.Called(data)
	return args.Get(0).(*NodeConfig), args.Error(0)
}

func (m MockService) CreateTenant(data config.TenantConfiguration) (config.TenantConfiguration, error) {
	args := m.Called(data)
	return args.Get(0).(*TenantConfig), args.Error(0)
}

func (m MockService) UpdateConfig(data config.Configuration) (config.Configuration, error) {
	args := m.Called()
	return args.Get(0).(*NodeConfig), args.Error(0)
}

func (m MockService) UpdateTenant(data config.TenantConfiguration) (config.TenantConfiguration, error) {
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
