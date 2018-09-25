// +build unit

package identity

import (
	"context"
	"fmt"
	"math/big"
	"testing"

	"github.com/centrifuge/go-centrifuge/centrifuge/tools"
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

func (i *mockID) GetCentrifugeID() CentID {
	args := i.Called()
	return args.Get(0).(CentID)
}

func (i *mockID) CentrifugeID(centId CentID) {
	i.Called(centId)
}

func (i *mockID) GetCurrentP2PKey() (ret string, err error) {
	args := i.Called()
	return args.String(0), args.Error(1)
}

func (i *mockID) GetLastKeyForPurpose(keyPurpose int) (key []byte, err error) {
	args := i.Called(keyPurpose)
	return args.Get(0).([]byte), args.Error(1)
}

func (i *mockID) AddKeyToIdentity(ctx context.Context, keyPurpose int, key []byte) (confirmations chan *WatchIdentity, err error) {
	args := i.Called(ctx, keyPurpose, key)
	return args.Get(0).(chan *WatchIdentity), args.Error(1)
}

func (i *mockID) CheckIdentityExists() (exists bool, err error) {
	args := i.Called()
	return args.Bool(0), args.Error(1)
}

func (i *mockID) FetchKey(key []byte) (Key, error) {
	args := i.Called(key)
	idKey, _ := args.Get(0).(Key)
	return idKey, args.Error(1)
}

// mockIDService implements Service
type mockIDService struct {
	mock.Mock
}

func (srv *mockIDService) LookupIdentityForID(centID CentID) (Identity, error) {
	args := srv.Called(centID)
	id, _ := args.Get(0).(Identity)
	return id, args.Error(1)
}

func (srv *mockIDService) CreateIdentity(centID CentID) (Identity, chan *WatchIdentity, error) {
	args := srv.Called(centID)
	id, _ := args.Get(0).(Identity)
	return id, args.Get(1).(chan *WatchIdentity), args.Error(2)
}

func (srv *mockIDService) CheckIdentityExists(centID CentID) (exists bool, err error) {
	args := srv.Called(centID)
	return args.Bool(0), args.Error(1)
}

func TestNewCentId(t *testing.T) {
	tests := []struct {
		name  string
		slice []byte
		err   string
	}{
		{
			"smallerSlice",
			tools.RandomSlice(CentIDByteLength - 1),
			"invalid length byte slice provided for centId",
		},
		{
			"largerSlice",
			tools.RandomSlice(CentIDByteLength + 1),
			"invalid length byte slice provided for centId",
		},
		{
			"nilSlice",
			nil,
			"invalid length byte slice provided for centId",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := NewCentID(test.slice)
			assert.Equal(t, test.err, err.Error())
		})
	}
}

func TestNewCentIdEqual(t *testing.T) {

	randomBytes := tools.RandomSlice(CentIDByteLength)

	centrifugeIdA, err := NewCentID(randomBytes)
	assert.Nil(t, err, "centrifugeId not initialized correctly ")

	centrifugeIdB, err := NewCentID(randomBytes)
	assert.Nil(t, err, "centrifugeId not initialized correctly ")

	assert.True(t, centrifugeIdA.Equal(centrifugeIdB), "centrifuge Id's should be the equal")

	randomBytes = tools.RandomSlice(CentIDByteLength)
	centrifugeIdC, _ := NewCentID(randomBytes)

	assert.False(t, centrifugeIdA.Equal(centrifugeIdC), "centrifuge Id's should not be equal")

}

func TestGetClientP2PURL_fail_service(t *testing.T) {
	centID, _ := NewCentID(tools.RandomSlice(CentIDByteLength))
	srv := &mockIDService{}
	srv.On("LookupIdentityForID", centID).Return(nil, fmt.Errorf("failed service")).Once()
	IDService = srv
	p2p, err := GetClientP2PURL(centID)
	srv.AssertExpectations(t)
	assert.Empty(t, p2p, "p2p is not empty")
	assert.Errorf(t, err, "error should not be nil")
	assert.Contains(t, err.Error(), "failed service")
}

func TestGetClientP2PURL_fail_identity(t *testing.T) {
	centID, _ := NewCentID(tools.RandomSlice(CentIDByteLength))
	srv := &mockIDService{}
	id := &mockID{}
	id.On("GetCurrentP2PKey").Return("", fmt.Errorf("error identity")).Once()
	srv.On("LookupIdentityForID", centID).Return(id, nil).Once()
	IDService = srv
	p2p, err := GetClientP2PURL(centID)
	srv.AssertExpectations(t)
	id.AssertExpectations(t)
	assert.Empty(t, p2p, "p2p is not empty")
	assert.Errorf(t, err, "error should not be nil")
	assert.Contains(t, err.Error(), "error identity")
}

func TestGetClientP2PURL_Success(t *testing.T) {
	centID, _ := NewCentID(tools.RandomSlice(CentIDByteLength))
	srv := &mockIDService{}
	id := &mockID{}
	id.On("GetCurrentP2PKey").Return("target", nil).Once()
	srv.On("LookupIdentityForID", centID).Return(id, nil).Once()
	IDService = srv
	p2p, err := GetClientP2PURL(centID)
	srv.AssertExpectations(t)
	id.AssertExpectations(t)
	assert.Nil(t, err, "must be nil")
	assert.Equal(t, p2p, "/ipfs/target")
}

func TestGetClientsP2PURLs_fail(t *testing.T) {
	centID, _ := NewCentID(tools.RandomSlice(CentIDByteLength))
	centIDs := []CentID{centID}
	srv := &mockIDService{}
	id := &mockID{}
	id.On("GetCurrentP2PKey").Return("", fmt.Errorf("error identity")).Once()
	srv.On("LookupIdentityForID", centIDs[0]).Return(id, nil).Once()
	IDService = srv
	p2pURLs, err := GetClientsP2PURLs(centIDs)
	srv.AssertExpectations(t)
	id.AssertExpectations(t)
	assert.Empty(t, p2pURLs, "p2p is not empty")
	assert.Errorf(t, err, "error should not be nil")
	assert.Contains(t, err.Error(), "error identity")
}

func TestGetClientsP2PURLs_success(t *testing.T) {
	centID, _ := NewCentID(tools.RandomSlice(CentIDByteLength))
	centIDs := []CentID{centID}
	id := &mockID{}
	id.On("GetCurrentP2PKey").Return("target", nil).Once()
	srv := &mockIDService{}
	srv.On("LookupIdentityForID", centIDs[0]).Return(id, nil).Once()
	IDService = srv
	p2pURLs, err := GetClientsP2PURLs(centIDs)
	srv.AssertExpectations(t)
	id.AssertExpectations(t)
	assert.Nil(t, err, "should be nil")
	assert.NotEmpty(t, p2pURLs, "should not be empty")
	assert.Equal(t, p2pURLs[0], "/ipfs/target")
}

func TestGetIdentityKey_fail_lookup(t *testing.T) {
	centID, _ := NewCentID(tools.RandomSlice(CentIDByteLength))
	srv := &mockIDService{}
	srv.On("LookupIdentityForID", centID).Return(nil, fmt.Errorf("lookup failed")).Once()
	IDService = srv
	id, err := GetIdentityKey(centID, tools.RandomSlice(32))
	srv.AssertExpectations(t)
	assert.Error(t, err, "must be not nil")
	assert.Contains(t, err.Error(), "lookup failed")
	assert.Nil(t, id, "must be nil")
}

func TestGetIdentityKey_fail_FetchKey(t *testing.T) {
	centID, _ := NewCentID(tools.RandomSlice(CentIDByteLength))
	pubKey := tools.RandomSlice(32)
	id := &mockID{}
	srv := &mockIDService{}
	srv.On("LookupIdentityForID", centID).Return(id, nil).Once()
	id.On("FetchKey", pubKey).Return(nil, fmt.Errorf("fetch key failed")).Once()
	IDService = srv
	key, err := GetIdentityKey(centID, pubKey)
	srv.AssertExpectations(t)
	id.AssertExpectations(t)
	assert.Error(t, err, "must be not nil")
	assert.Contains(t, err.Error(), "fetch key failed")
	assert.Nil(t, key, "must be nil")
}

func TestGetIdentityKey_fail_empty(t *testing.T) {
	centID, _ := NewCentID(tools.RandomSlice(CentIDByteLength))
	pubKey := tools.RandomSlice(32)
	var emptyKey [32]byte
	idkey := &EthereumIdentityKey{Key: emptyKey}
	id := &mockID{}
	srv := &mockIDService{}
	srv.On("LookupIdentityForID", centID).Return(id, nil).Once()
	id.On("FetchKey", pubKey).Return(idkey, nil).Once()
	IDService = srv
	key, err := GetIdentityKey(centID, pubKey)
	srv.AssertExpectations(t)
	id.AssertExpectations(t)
	assert.Error(t, err, "must be not nil")
	assert.Contains(t, err.Error(), "key not found for identity")
	assert.Nil(t, key, "must be nil")
}

func TestGetIdentityKey_Success(t *testing.T) {
	centID, _ := NewCentID(tools.RandomSlice(CentIDByteLength))
	pubKey := tools.RandomSlice(32)
	pkey := tools.RandomByte32()
	idkey := &EthereumIdentityKey{Key: pkey}
	id := &mockID{}
	srv := &mockIDService{}
	srv.On("LookupIdentityForID", centID).Return(id, nil).Once()
	id.On("FetchKey", pubKey).Return(idkey, nil).Once()
	IDService = srv
	key, err := GetIdentityKey(centID, pubKey)
	srv.AssertExpectations(t)
	id.AssertExpectations(t)
	assert.Nil(t, err, "error must be nil")
	assert.NotNil(t, key, "must not be nil")
	assert.Equal(t, key, idkey)
}

func TestValidateKey_fail_getId(t *testing.T) {
	centID, _ := NewCentID(tools.RandomSlice(CentIDByteLength))
	pubKey := tools.RandomSlice(32)
	srv := &mockIDService{}
	srv.On("LookupIdentityForID", centID).Return(nil, fmt.Errorf("failed at GetIdentity")).Once()
	IDService = srv
	err := ValidateKey(centID, pubKey, KeyPurposeSigning)
	srv.AssertExpectations(t)
	assert.Error(t, err, "must be not nil")
	assert.Contains(t, err.Error(), "failed at GetIdentity")
}

func TestValidateKey_fail_mismatch_key(t *testing.T) {
	centID, _ := NewCentID(tools.RandomSlice(CentIDByteLength))
	pubKey := tools.RandomSlice(32)
	idkey := &EthereumIdentityKey{Key: tools.RandomByte32()}
	id := &mockID{}
	srv := &mockIDService{}
	srv.On("LookupIdentityForID", centID).Return(id, nil).Once()
	id.On("FetchKey", pubKey).Return(idkey, nil).Once()
	IDService = srv
	err := ValidateKey(centID, pubKey, KeyPurposeSigning)
	srv.AssertExpectations(t)
	id.AssertExpectations(t)
	assert.Error(t, err, "must be not nil")
	assert.Contains(t, err.Error(), " Key doesn't match")
}

func TestValidateKey_fail_missing_purpose(t *testing.T) {
	centID, _ := NewCentID(tools.RandomSlice(CentIDByteLength))
	pubKey := tools.RandomByte32()
	idkey := &EthereumIdentityKey{Key: pubKey}
	id := &mockID{}
	srv := &mockIDService{}
	srv.On("LookupIdentityForID", centID).Return(id, nil).Once()
	id.On("FetchKey", pubKey[:]).Return(idkey, nil).Once()
	IDService = srv
	err := ValidateKey(centID, pubKey[:], KeyPurposeSigning)
	srv.AssertExpectations(t)
	id.AssertExpectations(t)
	assert.Error(t, err, "must be not nil")
	assert.Contains(t, err.Error(), "Key doesn't have purpose")
}

func TestValidateKey_fail_wrong_purpose(t *testing.T) {
	centID, _ := NewCentID(tools.RandomSlice(CentIDByteLength))
	pubKey := tools.RandomByte32()
	idkey := &EthereumIdentityKey{
		Key:      pubKey,
		Purposes: []*big.Int{big.NewInt(KeyPurposeEthMsgAuth)},
	}
	id := &mockID{}
	srv := &mockIDService{}
	srv.On("LookupIdentityForID", centID).Return(id, nil).Once()
	id.On("FetchKey", pubKey[:]).Return(idkey, nil).Once()
	IDService = srv
	err := ValidateKey(centID, pubKey[:], KeyPurposeSigning)
	srv.AssertExpectations(t)
	id.AssertExpectations(t)
	assert.Error(t, err, "must be not nil")
	assert.Contains(t, err.Error(), "Key doesn't have purpose")
}

func TestValidateKey_fail_revocation(t *testing.T) {
	centID, _ := NewCentID(tools.RandomSlice(CentIDByteLength))
	pubKey := tools.RandomByte32()
	idkey := &EthereumIdentityKey{
		Key:       pubKey,
		Purposes:  []*big.Int{big.NewInt(KeyPurposeSigning)},
		RevokedAt: big.NewInt(1),
	}
	id := &mockID{}
	srv := &mockIDService{}
	srv.On("LookupIdentityForID", centID).Return(id, nil).Once()
	id.On("FetchKey", pubKey[:]).Return(idkey, nil).Once()
	IDService = srv
	err := ValidateKey(centID, pubKey[:], KeyPurposeSigning)
	srv.AssertExpectations(t)
	id.AssertExpectations(t)
	assert.Error(t, err, "must be not nil")
	assert.Contains(t, err.Error(), "Key is currently revoked since block")
}

func TestValidateKey_success(t *testing.T) {
	centID, _ := NewCentID(tools.RandomSlice(CentIDByteLength))
	pubKey := tools.RandomByte32()
	idkey := &EthereumIdentityKey{
		Key:       pubKey,
		Purposes:  []*big.Int{big.NewInt(KeyPurposeSigning)},
		RevokedAt: big.NewInt(0),
	}
	id := &mockID{}
	srv := &mockIDService{}
	srv.On("LookupIdentityForID", centID).Return(id, nil).Once()
	id.On("FetchKey", pubKey[:]).Return(idkey, nil).Once()
	IDService = srv
	err := ValidateKey(centID, pubKey[:], KeyPurposeSigning)
	srv.AssertExpectations(t)
	id.AssertExpectations(t)
	assert.Nil(t, err, "must be nil")
}
