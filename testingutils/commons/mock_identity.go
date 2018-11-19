// +build unit integration

/*
The reason for this package is to avoid any kind of cyclic dependencies but share common mocking interfaces from identity packages across other packages
*/
package testingcommons

import (
	"context"

	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/mock"
)

// MockIDService implements Service
type MockIDService struct {
	mock.Mock
}

func (srv *MockIDService) ValidateSignature(signature *coredocumentpb.Signature, message []byte) error {
	args := srv.Called(signature, message)
	return args.Error(0)
}

func (srv *MockIDService) GetClientP2PURL(centID identity.CentID) (url string, err error) {
	args := srv.Called(centID)
	addr := args.Get(0).(string)
	return addr, args.Error(1)
}

func (srv *MockIDService) GetClientsP2PURLs(centIDs []identity.CentID) ([]string, error) {
	args := srv.Called(centIDs)
	addr := args.Get(0).([]string)
	return addr, args.Error(1)
}

func (srv *MockIDService) GetIdentityKey(id identity.CentID, pubKey []byte) (keyInfo identity.Key, err error) {
	args := srv.Called(id, pubKey)
	addr := args.Get(0).(identity.Key)
	return addr, args.Error(1)
}

func (srv *MockIDService) ValidateKey(centrifugeId identity.CentID, key []byte, purpose int) error {
	args := srv.Called(centrifugeId, key, purpose)
	return args.Error(0)
}

func (srv *MockIDService) AddKeyFromConfig(purpose int) error {
	args := srv.Called(purpose)
	return args.Error(0)
}

func (srv *MockIDService) GetIdentityAddress(centID identity.CentID) (common.Address, error) {
	args := srv.Called(centID)
	addr := args.Get(0).(common.Address)
	return addr, args.Error(1)
}

func (srv *MockIDService) LookupIdentityForID(centID identity.CentID) (identity.Identity, error) {
	args := srv.Called(centID)
	if id, ok := args.Get(0).(identity.Identity); ok {
		return id, args.Error(1)
	}
	return nil, args.Error(1)
}

func (srv *MockIDService) CreateIdentity(centID identity.CentID) (identity.Identity, chan *identity.WatchIdentity, error) {
	args := srv.Called(centID)
	id, _ := args.Get(0).(identity.Identity)
	watch, _ := args.Get(1).(chan *identity.WatchIdentity)
	return id, watch, args.Error(2)
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

func (i *MockID) CentID() identity.CentID {
	args := i.Called()
	return args.Get(0).(identity.CentID)
}

func (i *MockID) SetCentrifugeID(centId identity.CentID) {
	i.Called(centId)
}

func (i *MockID) CurrentP2PKey() (ret string, err error) {
	args := i.Called()
	return args.String(0), args.Error(1)
}

func (i *MockID) LastKeyForPurpose(keyPurpose int) (key []byte, err error) {
	args := i.Called(keyPurpose)
	return args.Get(0).([]byte), args.Error(1)
}

func (i *MockID) AddKeyToIdentity(ctx context.Context, keyPurpose int, key []byte) (confirmations chan *identity.WatchIdentity, err error) {
	args := i.Called(ctx, keyPurpose, key)
	return args.Get(0).(chan *identity.WatchIdentity), args.Error(1)
}

func (i *MockID) FetchKey(key []byte) (identity.Key, error) {
	args := i.Called(key)
	idKey, _ := args.Get(0).(identity.Key)
	return idKey, args.Error(1)
}
