/*
The reason for this package is to avoid any kind of cyclic dependencies but share common mocking interfaces from identity packages across other packages
*/
package testingcommons

import (
	"github.com/centrifuge/go-centrifuge/centrifuge/identity"
	"github.com/stretchr/testify/mock"
)

// MockIDService implements Service
type MockIDService struct {
	mock.Mock
}

func (srv *MockIDService) LookupIdentityForID(centID identity.CentID) (identity.Identity, error) {
	args := srv.Called(centID)
	id, _ := args.Get(0).(identity.Identity)
	return id, args.Error(1)
}

func (srv *MockIDService) CreateIdentity(centID identity.CentID) (identity.Identity, chan *identity.WatchIdentity, error) {
	args := srv.Called(centID)
	id, _ := args.Get(0).(identity.Identity)
	return id, args.Get(1).(chan *identity.WatchIdentity), args.Error(2)
}

func (srv *MockIDService) CheckIdentityExists(centID identity.CentID) (exists bool, err error) {
	args := srv.Called(centID)
	return args.Bool(0), args.Error(1)
}

// MockID implements Identity
type MockID struct {
	mock.Mock
}

func (i *MockID) String() string {
	args := i.Called()
	return args.String(0)
}

func (i *MockID) GetCentrifugeID() identity.CentID {
	args := i.Called()
	return args.Get(0).(identity.CentID)
}

func (i *MockID) CentrifugeID(centId identity.CentID) {
	i.Called(centId)
}

func (i *MockID) GetCurrentP2PKey() (ret string, err error) {
	args := i.Called()
	return args.String(0), args.Error(1)
}

func (i *MockID) GetLastKeyForPurpose(keyPurpose int) (key []byte, err error) {
	args := i.Called(keyPurpose)
	return args.Get(0).([]byte), args.Error(1)
}

func (i *MockID) AddKeyToIdentity(keyPurpose int, key []byte) (confirmations chan *identity.WatchIdentity, err error) {
	args := i.Called(keyPurpose, key)
	return args.Get(0).(chan *identity.WatchIdentity), args.Error(1)
}

func (i *MockID) CheckIdentityExists() (exists bool, err error) {
	args := i.Called()
	return args.Bool(0), args.Error(1)
}

func (i *MockID) FetchKey(key []byte) (identity.Key, error) {
	args := i.Called(key)
	idKey, _ := args.Get(0).(identity.Key)
	return idKey, args.Error(1)
}
