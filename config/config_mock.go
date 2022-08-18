// Code generated by mockery v2.13.0-beta.1. DO NOT EDIT.

package config

import (
	reflect "reflect"

	mock "github.com/stretchr/testify/mock"

	time "time"
)

// ConfigurationMock is an autogenerated mock type for the Configuration type
type ConfigurationMock struct {
	mock.Mock
}

// FromJSON provides a mock function with given fields: json
func (_m *ConfigurationMock) FromJSON(json []byte) error {
	ret := _m.Called(json)

	var r0 error
	if rf, ok := ret.Get(0).(func([]byte) error); ok {
		r0 = rf(json)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// GetBootstrapPeers provides a mock function with given fields:
func (_m *ConfigurationMock) GetBootstrapPeers() []string {
	ret := _m.Called()

	var r0 []string
	if rf, ok := ret.Get(0).(func() []string); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]string)
		}
	}

	return r0
}

// GetCentChainAnchorLifespan provides a mock function with given fields:
func (_m *ConfigurationMock) GetCentChainAnchorLifespan() time.Duration {
	ret := _m.Called()

	var r0 time.Duration
	if rf, ok := ret.Get(0).(func() time.Duration); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(time.Duration)
	}

	return r0
}

// GetCentChainIntervalRetry provides a mock function with given fields:
func (_m *ConfigurationMock) GetCentChainIntervalRetry() time.Duration {
	ret := _m.Called()

	var r0 time.Duration
	if rf, ok := ret.Get(0).(func() time.Duration); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(time.Duration)
	}

	return r0
}

// GetCentChainMaxRetries provides a mock function with given fields:
func (_m *ConfigurationMock) GetCentChainMaxRetries() int {
	ret := _m.Called()

	var r0 int
	if rf, ok := ret.Get(0).(func() int); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(int)
	}

	return r0
}

// GetCentChainNodeURL provides a mock function with given fields:
func (_m *ConfigurationMock) GetCentChainNodeURL() string {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// GetConfigStoragePath provides a mock function with given fields:
func (_m *ConfigurationMock) GetConfigStoragePath() string {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// GetIPFSPinningServiceAuth provides a mock function with given fields:
func (_m *ConfigurationMock) GetIPFSPinningServiceAuth() string {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// GetIPFSPinningServiceName provides a mock function with given fields:
func (_m *ConfigurationMock) GetIPFSPinningServiceName() string {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// GetIPFSPinningServiceURL provides a mock function with given fields:
func (_m *ConfigurationMock) GetIPFSPinningServiceURL() string {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// GetNetworkID provides a mock function with given fields:
func (_m *ConfigurationMock) GetNetworkID() uint32 {
	ret := _m.Called()

	var r0 uint32
	if rf, ok := ret.Get(0).(func() uint32); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(uint32)
	}

	return r0
}

// GetNetworkString provides a mock function with given fields:
func (_m *ConfigurationMock) GetNetworkString() string {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// GetNodeAdminKeyPair provides a mock function with given fields:
func (_m *ConfigurationMock) GetNodeAdminKeyPair() (string, string) {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	var r1 string
	if rf, ok := ret.Get(1).(func() string); ok {
		r1 = rf()
	} else {
		r1 = ret.Get(1).(string)
	}

	return r0, r1
}

// GetNumWorkers provides a mock function with given fields:
func (_m *ConfigurationMock) GetNumWorkers() int {
	ret := _m.Called()

	var r0 int
	if rf, ok := ret.Get(0).(func() int); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(int)
	}

	return r0
}

// GetP2PConnectionTimeout provides a mock function with given fields:
func (_m *ConfigurationMock) GetP2PConnectionTimeout() time.Duration {
	ret := _m.Called()

	var r0 time.Duration
	if rf, ok := ret.Get(0).(func() time.Duration); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(time.Duration)
	}

	return r0
}

// GetP2PExternalIP provides a mock function with given fields:
func (_m *ConfigurationMock) GetP2PExternalIP() string {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// GetP2PKeyPair provides a mock function with given fields:
func (_m *ConfigurationMock) GetP2PKeyPair() (string, string) {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	var r1 string
	if rf, ok := ret.Get(1).(func() string); ok {
		r1 = rf()
	} else {
		r1 = ret.Get(1).(string)
	}

	return r0, r1
}

// GetP2PPort provides a mock function with given fields:
func (_m *ConfigurationMock) GetP2PPort() int {
	ret := _m.Called()

	var r0 int
	if rf, ok := ret.Get(0).(func() int); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(int)
	}

	return r0
}

// GetP2PResponseDelay provides a mock function with given fields:
func (_m *ConfigurationMock) GetP2PResponseDelay() time.Duration {
	ret := _m.Called()

	var r0 time.Duration
	if rf, ok := ret.Get(0).(func() time.Duration); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(time.Duration)
	}

	return r0
}

// GetServerAddress provides a mock function with given fields:
func (_m *ConfigurationMock) GetServerAddress() string {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// GetServerPort provides a mock function with given fields:
func (_m *ConfigurationMock) GetServerPort() int {
	ret := _m.Called()

	var r0 int
	if rf, ok := ret.Get(0).(func() int); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(int)
	}

	return r0
}

// GetSigningKeyPair provides a mock function with given fields:
func (_m *ConfigurationMock) GetSigningKeyPair() (string, string) {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	var r1 string
	if rf, ok := ret.Get(1).(func() string); ok {
		r1 = rf()
	} else {
		r1 = ret.Get(1).(string)
	}

	return r0, r1
}

// GetStoragePath provides a mock function with given fields:
func (_m *ConfigurationMock) GetStoragePath() string {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// GetTaskValidDuration provides a mock function with given fields:
func (_m *ConfigurationMock) GetTaskValidDuration() time.Duration {
	ret := _m.Called()

	var r0 time.Duration
	if rf, ok := ret.Get(0).(func() time.Duration); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(time.Duration)
	}

	return r0
}

// GetWorkerWaitTimeMS provides a mock function with given fields:
func (_m *ConfigurationMock) GetWorkerWaitTimeMS() int {
	ret := _m.Called()

	var r0 int
	if rf, ok := ret.Get(0).(func() int); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(int)
	}

	return r0
}

// IsAuthenticationEnabled provides a mock function with given fields:
func (_m *ConfigurationMock) IsAuthenticationEnabled() bool {
	ret := _m.Called()

	var r0 bool
	if rf, ok := ret.Get(0).(func() bool); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}

// IsDebugLogEnabled provides a mock function with given fields:
func (_m *ConfigurationMock) IsDebugLogEnabled() bool {
	ret := _m.Called()

	var r0 bool
	if rf, ok := ret.Get(0).(func() bool); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}

// IsPProfEnabled provides a mock function with given fields:
func (_m *ConfigurationMock) IsPProfEnabled() bool {
	ret := _m.Called()

	var r0 bool
	if rf, ok := ret.Get(0).(func() bool); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}

// JSON provides a mock function with given fields:
func (_m *ConfigurationMock) JSON() ([]byte, error) {
	ret := _m.Called()

	var r0 []byte
	if rf, ok := ret.Get(0).(func() []byte); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]byte)
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

// Type provides a mock function with given fields:
func (_m *ConfigurationMock) Type() reflect.Type {
	ret := _m.Called()

	var r0 reflect.Type
	if rf, ok := ret.Get(0).(func() reflect.Type); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(reflect.Type)
		}
	}

	return r0
}

type NewConfigurationMockT interface {
	mock.TestingT
	Cleanup(func())
}

// NewConfigurationMock creates a new instance of ConfigurationMock. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewConfigurationMock(t NewConfigurationMockT) *ConfigurationMock {
	mock := &ConfigurationMock{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
