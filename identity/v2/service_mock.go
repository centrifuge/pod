// Code generated by mockery v2.13.0-beta.1. DO NOT EDIT.

package v2

import (
	context "context"

	config "github.com/centrifuge/pod/config"

	keystore "github.com/centrifuge/chain-custom-types/pkg/keystore"

	mock "github.com/stretchr/testify/mock"

	time "time"

	types "github.com/centrifuge/go-substrate-rpc-client/v4/types"
)

// ServiceMock is an autogenerated mock type for the Service type
type ServiceMock struct {
	mock.Mock
}

// CreateIdentity provides a mock function with given fields: ctx, req
func (_m *ServiceMock) CreateIdentity(ctx context.Context, req *CreateIdentityRequest) (config.Account, error) {
	ret := _m.Called(ctx, req)

	var r0 config.Account
	if rf, ok := ret.Get(0).(func(context.Context, *CreateIdentityRequest) config.Account); ok {
		r0 = rf(ctx, req)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(config.Account)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *CreateIdentityRequest) error); ok {
		r1 = rf(ctx, req)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ValidateAccount provides a mock function with given fields: accountID
func (_m *ServiceMock) ValidateAccount(accountID *types.AccountID) error {
	ret := _m.Called(accountID)

	var r0 error
	if rf, ok := ret.Get(0).(func(*types.AccountID) error); ok {
		r0 = rf(accountID)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// ValidateDocumentSignature provides a mock function with given fields: accountID, pubKey, message, signature, validationTime
func (_m *ServiceMock) ValidateDocumentSignature(accountID *types.AccountID, pubKey []byte, message []byte, signature []byte, validationTime time.Time) error {
	ret := _m.Called(accountID, pubKey, message, signature, validationTime)

	var r0 error
	if rf, ok := ret.Get(0).(func(*types.AccountID, []byte, []byte, []byte, time.Time) error); ok {
		r0 = rf(accountID, pubKey, message, signature, validationTime)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// ValidateKey provides a mock function with given fields: accountID, pubKey, keyPurpose, validationTime
func (_m *ServiceMock) ValidateKey(accountID *types.AccountID, pubKey []byte, keyPurpose keystore.KeyPurpose, validationTime time.Time) error {
	ret := _m.Called(accountID, pubKey, keyPurpose, validationTime)

	var r0 error
	if rf, ok := ret.Get(0).(func(*types.AccountID, []byte, keystore.KeyPurpose, time.Time) error); ok {
		r0 = rf(accountID, pubKey, keyPurpose, validationTime)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

type NewServiceMockT interface {
	mock.TestingT
	Cleanup(func())
}

// NewServiceMock creates a new instance of ServiceMock. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewServiceMock(t NewServiceMockT) *ServiceMock {
	mock := &ServiceMock{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
