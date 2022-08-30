// Code generated by mockery v2.13.0-beta.1. DO NOT EDIT.

package configstore

import (
	config "github.com/centrifuge/go-centrifuge/config"
	mock "github.com/stretchr/testify/mock"
)

// RepositoryMock is an autogenerated mock type for the Repository type
type RepositoryMock struct {
	mock.Mock
}

// CreateAccount provides a mock function with given fields: acc
func (_m *RepositoryMock) CreateAccount(acc config.Account) error {
	ret := _m.Called(acc)

	var r0 error
	if rf, ok := ret.Get(0).(func(config.Account) error); ok {
		r0 = rf(acc)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// CreateConfig provides a mock function with given fields: cfg
func (_m *RepositoryMock) CreateConfig(cfg config.Configuration) error {
	ret := _m.Called(cfg)

	var r0 error
	if rf, ok := ret.Get(0).(func(config.Configuration) error); ok {
		r0 = rf(cfg)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// CreateNodeAdmin provides a mock function with given fields: nodeAdmin
func (_m *RepositoryMock) CreateNodeAdmin(nodeAdmin config.NodeAdmin) error {
	ret := _m.Called(nodeAdmin)

	var r0 error
	if rf, ok := ret.Get(0).(func(config.NodeAdmin) error); ok {
		r0 = rf(nodeAdmin)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// CreatePodOperator provides a mock function with given fields: podOperator
func (_m *RepositoryMock) CreatePodOperator(podOperator config.PodOperator) error {
	ret := _m.Called(podOperator)

	var r0 error
	if rf, ok := ret.Get(0).(func(config.PodOperator) error); ok {
		r0 = rf(podOperator)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// DeleteAccount provides a mock function with given fields: id
func (_m *RepositoryMock) DeleteAccount(id []byte) error {
	ret := _m.Called(id)

	var r0 error
	if rf, ok := ret.Get(0).(func([]byte) error); ok {
		r0 = rf(id)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// DeleteConfig provides a mock function with given fields:
func (_m *RepositoryMock) DeleteConfig() error {
	ret := _m.Called()

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// GetAccount provides a mock function with given fields: id
func (_m *RepositoryMock) GetAccount(id []byte) (config.Account, error) {
	ret := _m.Called(id)

	var r0 config.Account
	if rf, ok := ret.Get(0).(func([]byte) config.Account); ok {
		r0 = rf(id)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(config.Account)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func([]byte) error); ok {
		r1 = rf(id)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetAllAccounts provides a mock function with given fields:
func (_m *RepositoryMock) GetAllAccounts() ([]config.Account, error) {
	ret := _m.Called()

	var r0 []config.Account
	if rf, ok := ret.Get(0).(func() []config.Account); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]config.Account)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetConfig provides a mock function with given fields:
func (_m *RepositoryMock) GetConfig() (config.Configuration, error) {
	ret := _m.Called()

	var r0 config.Configuration
	if rf, ok := ret.Get(0).(func() config.Configuration); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(config.Configuration)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetNodeAdmin provides a mock function with given fields:
func (_m *RepositoryMock) GetNodeAdmin() (config.NodeAdmin, error) {
	ret := _m.Called()

	var r0 config.NodeAdmin
	if rf, ok := ret.Get(0).(func() config.NodeAdmin); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(config.NodeAdmin)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetPodOperator provides a mock function with given fields:
func (_m *RepositoryMock) GetPodOperator() (config.PodOperator, error) {
	ret := _m.Called()

	var r0 config.PodOperator
	if rf, ok := ret.Get(0).(func() config.PodOperator); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(config.PodOperator)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// RegisterAccount provides a mock function with given fields: acc
func (_m *RepositoryMock) RegisterAccount(acc config.Account) {
	_m.Called(acc)
}

// RegisterConfig provides a mock function with given fields: cfg
func (_m *RepositoryMock) RegisterConfig(cfg config.Configuration) {
	_m.Called(cfg)
}

// RegisterNodeAdmin provides a mock function with given fields: nodeAdmin
func (_m *RepositoryMock) RegisterNodeAdmin(nodeAdmin config.NodeAdmin) {
	_m.Called(nodeAdmin)
}

// RegisterPodOperator provides a mock function with given fields: podOperator
func (_m *RepositoryMock) RegisterPodOperator(podOperator config.PodOperator) {
	_m.Called(podOperator)
}

// UpdateAccount provides a mock function with given fields: acc
func (_m *RepositoryMock) UpdateAccount(acc config.Account) error {
	ret := _m.Called(acc)

	var r0 error
	if rf, ok := ret.Get(0).(func(config.Account) error); ok {
		r0 = rf(acc)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// UpdateConfig provides a mock function with given fields: cfg
func (_m *RepositoryMock) UpdateConfig(cfg config.Configuration) error {
	ret := _m.Called(cfg)

	var r0 error
	if rf, ok := ret.Get(0).(func(config.Configuration) error); ok {
		r0 = rf(cfg)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// UpdateNodeAdmin provides a mock function with given fields: nodeAdmin
func (_m *RepositoryMock) UpdateNodeAdmin(nodeAdmin config.NodeAdmin) error {
	ret := _m.Called(nodeAdmin)

	var r0 error
	if rf, ok := ret.Get(0).(func(config.NodeAdmin) error); ok {
		r0 = rf(nodeAdmin)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// UpdatePodOperator provides a mock function with given fields: podOperator
func (_m *RepositoryMock) UpdatePodOperator(podOperator config.PodOperator) error {
	ret := _m.Called(podOperator)

	var r0 error
	if rf, ok := ret.Get(0).(func(config.PodOperator) error); ok {
		r0 = rf(podOperator)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

type NewRepositoryMockT interface {
	mock.TestingT
	Cleanup(func())
}

// NewRepositoryMock creates a new instance of RepositoryMock. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewRepositoryMock(t NewRepositoryMockT) *RepositoryMock {
	mock := &RepositoryMock{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
