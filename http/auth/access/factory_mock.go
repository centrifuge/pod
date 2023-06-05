// Code generated by mockery v2.13.0-beta.1. DO NOT EDIT.

package access

import mock "github.com/stretchr/testify/mock"

// ValidationWrapperFactoryMock is an autogenerated mock type for the ValidationWrapperFactory type
type ValidationWrapperFactoryMock struct {
	mock.Mock
}

// GetValidationWrappers provides a mock function with given fields:
func (_m *ValidationWrapperFactoryMock) GetValidationWrappers() (ValidationWrappers, error) {
	ret := _m.Called()

	var r0 ValidationWrappers
	if rf, ok := ret.Get(0).(func() ValidationWrappers); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(ValidationWrappers)
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

type NewValidationWrapperFactoryMockT interface {
	mock.TestingT
	Cleanup(func())
}

// NewValidationWrapperFactoryMock creates a new instance of ValidationWrapperFactoryMock. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewValidationWrapperFactoryMock(t NewValidationWrapperFactoryMockT) *ValidationWrapperFactoryMock {
	mock := &ValidationWrapperFactoryMock{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
