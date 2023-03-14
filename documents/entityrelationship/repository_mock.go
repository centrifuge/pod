// Code generated by mockery v2.13.0-beta.1. DO NOT EDIT.

package entityrelationship

import (
	documents "github.com/centrifuge/pod/documents"
	mock "github.com/stretchr/testify/mock"

	types "github.com/centrifuge/go-substrate-rpc-client/v4/types"
)

// repositoryMock is an autogenerated mock type for the repository type
type repositoryMock struct {
	mock.Mock
}

// Create provides a mock function with given fields: accountID, id, model
func (_m *repositoryMock) Create(accountID []byte, id []byte, model documents.Document) error {
	ret := _m.Called(accountID, id, model)

	var r0 error
	if rf, ok := ret.Get(0).(func([]byte, []byte, documents.Document) error); ok {
		r0 = rf(accountID, id, model)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Exists provides a mock function with given fields: accountID, id
func (_m *repositoryMock) Exists(accountID []byte, id []byte) bool {
	ret := _m.Called(accountID, id)

	var r0 bool
	if rf, ok := ret.Get(0).(func([]byte, []byte) bool); ok {
		r0 = rf(accountID, id)
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}

// FindEntityRelationshipIdentifier provides a mock function with given fields: entityIdentifier, ownerAccountID, targetAccountID
func (_m *repositoryMock) FindEntityRelationshipIdentifier(entityIdentifier []byte, ownerAccountID *types.AccountID, targetAccountID *types.AccountID) ([]byte, error) {
	ret := _m.Called(entityIdentifier, ownerAccountID, targetAccountID)

	var r0 []byte
	if rf, ok := ret.Get(0).(func([]byte, *types.AccountID, *types.AccountID) []byte); ok {
		r0 = rf(entityIdentifier, ownerAccountID, targetAccountID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]byte)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func([]byte, *types.AccountID, *types.AccountID) error); ok {
		r1 = rf(entityIdentifier, ownerAccountID, targetAccountID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Get provides a mock function with given fields: accountID, id
func (_m *repositoryMock) Get(accountID []byte, id []byte) (documents.Document, error) {
	ret := _m.Called(accountID, id)

	var r0 documents.Document
	if rf, ok := ret.Get(0).(func([]byte, []byte) documents.Document); ok {
		r0 = rf(accountID, id)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(documents.Document)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func([]byte, []byte) error); ok {
		r1 = rf(accountID, id)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetLatest provides a mock function with given fields: accountID, docID
func (_m *repositoryMock) GetLatest(accountID []byte, docID []byte) (documents.Document, error) {
	ret := _m.Called(accountID, docID)

	var r0 documents.Document
	if rf, ok := ret.Get(0).(func([]byte, []byte) documents.Document); ok {
		r0 = rf(accountID, docID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(documents.Document)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func([]byte, []byte) error); ok {
		r1 = rf(accountID, docID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ListAllRelationships provides a mock function with given fields: entityIdentifier, ownerAccountID
func (_m *repositoryMock) ListAllRelationships(entityIdentifier []byte, ownerAccountID *types.AccountID) (map[string][]byte, error) {
	ret := _m.Called(entityIdentifier, ownerAccountID)

	var r0 map[string][]byte
	if rf, ok := ret.Get(0).(func([]byte, *types.AccountID) map[string][]byte); ok {
		r0 = rf(entityIdentifier, ownerAccountID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(map[string][]byte)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func([]byte, *types.AccountID) error); ok {
		r1 = rf(entityIdentifier, ownerAccountID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Register provides a mock function with given fields: model
func (_m *repositoryMock) Register(model documents.Document) {
	_m.Called(model)
}

// Update provides a mock function with given fields: accountID, id, model
func (_m *repositoryMock) Update(accountID []byte, id []byte, model documents.Document) error {
	ret := _m.Called(accountID, id, model)

	var r0 error
	if rf, ok := ret.Get(0).(func([]byte, []byte, documents.Document) error); ok {
		r0 = rf(accountID, id, model)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

type newRepositoryMockT interface {
	mock.TestingT
	Cleanup(func())
}

// newRepositoryMock creates a new instance of repositoryMock. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func newRepositoryMock(t newRepositoryMockT) *repositoryMock {
	mock := &repositoryMock{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
