// +build unit

package identity

import (
	"math/big"
	"testing"

	"fmt"

	"github.com/CentrifugeInc/go-centrifuge/centrifuge/testingutils"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/tools"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// mockID implements Identity
type mockID struct {
	mock.Mock
}

func (i *mockID) String() string {
	args := i.Called()
	return args.String(0)
}

func (i *mockID) GetCentrifugeID() []byte {
	args := i.Called()
	return args.Get(0).([]byte)
}

func (i *mockID) CentrifugeIDString() string {
	args := i.Called()
	return args.String(0)
}

func (i *mockID) CentrifugeIDBytes() [CentIdByteLength]byte {
	args := i.Called()
	return args.Get(0).([CentIdByteLength]byte)
}

func (i *mockID) CentrifugeIDBigInt() *big.Int {
	args := i.Called()
	return args.Get(0).(*big.Int)
}

func (i *mockID) SetCentrifugeID(b []byte) error {
	args := i.Called(b)
	return args.Error(0)
}

func (i *mockID) GetCurrentP2PKey() (ret string, err error) {
	args := i.Called()
	return args.String(0), args.Error(1)
}

func (i *mockID) GetLastKeyForPurpose(keyPurpose int) (key []byte, err error) {
	args := i.Called(keyPurpose)
	return args.Get(0).([]byte), args.Error(1)
}

func (i *mockID) AddKeyToIdentity(keyPurpose int, key []byte) (confirmations chan *WatchIdentity, err error) {
	args := i.Called(keyPurpose, key)
	return args.Get(0).(chan *WatchIdentity), args.Error(1)
}

func (i *mockID) CheckIdentityExists() (exists bool, err error) {
	args := i.Called()
	return args.Bool(0), args.Error(1)
}

func TestGetClientP2PURL_fail_service(t *testing.T) {
	centID := tools.RandomSlice(CentIdByteLength)
	srv := &testingutils.MockIDService{}
	srv.On("LookupIdentityForId", centID).Return(nil, fmt.Errorf("failed service")).Once()
	p2p, err := GetClientP2PURL(srv, centID)
	srv.AssertExpectations(t)
	assert.Empty(t, p2p, "p2p is not empty")
	assert.Errorf(t, err, "error should not be nil")
	assert.Contains(t, err.Error(), "failed service")
}

func TestGetClientP2PURL_fail_identity(t *testing.T) {
	centID := tools.RandomSlice(CentIdByteLength)
	srv := &testingutils.MockIDService{}
	id := &mockID{}
	id.On("GetCurrentP2PKey").Return("", fmt.Errorf("error identity")).Once()
	srv.On("LookupIdentityForId", centID).Return(id, nil).Once()
	p2p, err := GetClientP2PURL(srv, centID)
	srv.AssertExpectations(t)
	id.AssertExpectations(t)
	assert.Empty(t, p2p, "p2p is not empty")
	assert.Errorf(t, err, "error should not be nil")
	assert.Contains(t, err.Error(), "error identity")
}

func TestGetClientP2PURL_Success(t *testing.T) {
	centID := tools.RandomSlice(CentIdByteLength)
	srv := &testingutils.MockIDService{}
	id := &mockID{}
	id.On("GetCurrentP2PKey").Return("target", nil).Once()
	srv.On("LookupIdentityForId", centID).Return(id, nil).Once()
	p2p, err := GetClientP2PURL(srv, centID)
	srv.AssertExpectations(t)
	id.AssertExpectations(t)
	assert.Nil(t, err, "must be nil")
	assert.Equal(t, p2p, "/ipfs/target")
}

func TestGetClientsP2PURLs_fail(t *testing.T) {
	centIDs := [][]byte{tools.RandomSlice(CentIdByteLength)}
	srv := &testingutils.MockIDService{}
	id := &mockID{}
	id.On("GetCurrentP2PKey").Return("", fmt.Errorf("error identity")).Once()
	srv.On("LookupIdentityForId", centIDs[0]).Return(id, nil).Once()
	p2pURLs, err := GetClientsP2PURLs(srv, centIDs)
	srv.AssertExpectations(t)
	id.AssertExpectations(t)
	assert.Empty(t, p2pURLs, "p2p is not empty")
	assert.Errorf(t, err, "error should not be nil")
	assert.Contains(t, err.Error(), "error identity")
}

func TestGetClientsP2PURLs_success(t *testing.T) {
	centIDs := [][]byte{tools.RandomSlice(CentIdByteLength)}
	id := &mockID{}
	id.On("GetCurrentP2PKey").Return("target", nil).Once()
	srv := &testingutils.MockIDService{}
	srv.On("LookupIdentityForId", centIDs[0]).Return(id, nil).Once()
	p2pURLs, err := GetClientsP2PURLs(srv, centIDs)
	srv.AssertExpectations(t)
	id.AssertExpectations(t)
	assert.Nil(t, err, "should be nil")
	assert.NotEmpty(t, p2pURLs, "should not be empty")
	assert.Equal(t, p2pURLs[0], "/ipfs/target")
}
