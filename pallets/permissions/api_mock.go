// Code generated by mockery v2.13.0-beta.1. DO NOT EDIT.

package permissions

import (
	types "github.com/centrifuge/go-substrate-rpc-client/v4/types"
	mock "github.com/stretchr/testify/mock"
)

// APIMock is an autogenerated mock type for the API type
type APIMock struct {
	mock.Mock
}

// GetPermissionRoles provides a mock function with given fields: accountID, poolID
func (_m *APIMock) GetPermissionRoles(accountID *types.AccountID, poolID types.U64) (*PermissionRoles, error) {
	ret := _m.Called(accountID, poolID)

	var r0 *PermissionRoles
	if rf, ok := ret.Get(0).(func(*types.AccountID, types.U64) *PermissionRoles); ok {
		r0 = rf(accountID, poolID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*PermissionRoles)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(*types.AccountID, types.U64) error); ok {
		r1 = rf(accountID, poolID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

type NewAPIMockT interface {
	mock.TestingT
	Cleanup(func())
}

// NewAPIMock creates a new instance of APIMock. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewAPIMock(t NewAPIMockT) *APIMock {
	mock := &APIMock{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
