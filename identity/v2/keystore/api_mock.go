// Code generated by mockery v2.13.0-beta.1. DO NOT EDIT.

package keystore

import (
	context "context"

	centchain "github.com/centrifuge/go-centrifuge/centchain"

	mock "github.com/stretchr/testify/mock"

	pkgkeystore "github.com/centrifuge/chain-custom-types/pkg/keystore"

	types "github.com/centrifuge/go-substrate-rpc-client/v4/types"
)

// KeystoreAPIMock is an autogenerated mock type for the API type
type KeystoreAPIMock struct {
	mock.Mock
}

// AddKeys provides a mock function with given fields: ctx, keys
func (_m *KeystoreAPIMock) AddKeys(ctx context.Context, keys []*pkgkeystore.AddKey) (*centchain.ExtrinsicInfo, error) {
	ret := _m.Called(ctx, keys)

	var r0 *centchain.ExtrinsicInfo
	if rf, ok := ret.Get(0).(func(context.Context, []*pkgkeystore.AddKey) *centchain.ExtrinsicInfo); ok {
		r0 = rf(ctx, keys)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*centchain.ExtrinsicInfo)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, []*pkgkeystore.AddKey) error); ok {
		r1 = rf(ctx, keys)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetKey provides a mock function with given fields: ctx, keyID
func (_m *KeystoreAPIMock) GetKey(ctx context.Context, keyID *pkgkeystore.KeyID) (*pkgkeystore.Key, error) {
	ret := _m.Called(ctx, keyID)

	var r0 *pkgkeystore.Key
	if rf, ok := ret.Get(0).(func(context.Context, *pkgkeystore.KeyID) *pkgkeystore.Key); ok {
		r0 = rf(ctx, keyID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*pkgkeystore.Key)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *pkgkeystore.KeyID) error); ok {
		r1 = rf(ctx, keyID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetLastKeyByPurpose provides a mock function with given fields: ctx, keyPurpose
func (_m *KeystoreAPIMock) GetLastKeyByPurpose(ctx context.Context, keyPurpose pkgkeystore.KeyPurpose) (*types.Hash, error) {
	ret := _m.Called(ctx, keyPurpose)

	var r0 *types.Hash
	if rf, ok := ret.Get(0).(func(context.Context, pkgkeystore.KeyPurpose) *types.Hash); ok {
		r0 = rf(ctx, keyPurpose)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*types.Hash)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, pkgkeystore.KeyPurpose) error); ok {
		r1 = rf(ctx, keyPurpose)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// RevokeKeys provides a mock function with given fields: ctx, keys, keyPurpose
func (_m *KeystoreAPIMock) RevokeKeys(ctx context.Context, keys []*types.Hash, keyPurpose pkgkeystore.KeyPurpose) (*centchain.ExtrinsicInfo, error) {
	ret := _m.Called(ctx, keys, keyPurpose)

	var r0 *centchain.ExtrinsicInfo
	if rf, ok := ret.Get(0).(func(context.Context, []*types.Hash, pkgkeystore.KeyPurpose) *centchain.ExtrinsicInfo); ok {
		r0 = rf(ctx, keys, keyPurpose)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*centchain.ExtrinsicInfo)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, []*types.Hash, pkgkeystore.KeyPurpose) error); ok {
		r1 = rf(ctx, keys, keyPurpose)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

type NewKeystoreAPIMockT interface {
	mock.TestingT
	Cleanup(func())
}

// NewKeystoreAPIMock creates a new instance of KeystoreAPIMock. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewKeystoreAPIMock(t NewKeystoreAPIMockT) *KeystoreAPIMock {
	mock := &KeystoreAPIMock{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
