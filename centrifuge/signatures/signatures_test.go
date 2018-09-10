// +build unit

package signatures

import (
	"fmt"
	"testing"

	"github.com/CentrifugeInc/centrifuge-protobufs/gen/go/coredocument"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/config"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/identity"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/testingutils"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/tools"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var (
	key1Pub   = []byte{230, 49, 10, 12, 200, 149, 43, 184, 145, 87, 163, 252, 114, 31, 91, 163, 24, 237, 36, 51, 165, 8, 34, 104, 97, 49, 114, 85, 255, 15, 195, 199}
	key1      = []byte{102, 109, 71, 239, 130, 229, 128, 189, 37, 96, 223, 5, 189, 91, 210, 47, 89, 4, 165, 6, 188, 53, 49, 250, 109, 151, 234, 139, 57, 205, 231, 253, 230, 49, 10, 12, 200, 149, 43, 184, 145, 87, 163, 252, 114, 31, 91, 163, 24, 237, 36, 51, 165, 8, 34, 104, 97, 49, 114, 85, 255, 15, 195, 199}
	id1       = []byte{1, 1, 1, 1, 1, 1}
	signature = []byte{0x4e, 0x3d, 0x90, 0x5f, 0x25, 0xc7, 0x90, 0x63, 0x7e, 0x6c, 0xd0, 0xe6, 0xc7, 0xbd, 0xe6, 0x81, 0x3b, 0xd0, 0x5b, 0x94, 0x76, 0x86, 0x4e, 0xcb, 0xb9, 0x36, 0x48, 0x44, 0x4b, 0x98, 0xd2, 0x4b, 0x6a, 0x65, 0x22, 0x92, 0x1c, 0x8a, 0xdb, 0xfe, 0xb7, 0x6f, 0xfe, 0x34, 0x52, 0xa3, 0x49, 0xe4, 0xda, 0xdc, 0x5d, 0x1b, 0x0, 0x79, 0x54, 0x60, 0x29, 0x22, 0x94, 0xb, 0x3c, 0x90, 0x3c, 0x3}
)

func TestSign(t *testing.T) {
	coreDoc := testingutils.GenerateCoreDocument()
	coreDoc.SigningRoot = key1Pub
	idConfig := config.IdentityConfig{
		ID:         id1,
		PublicKey:  key1Pub,
		PrivateKey: key1,
	}

	sig := Sign(&idConfig, coreDoc)
	assert.NotNil(t, sig)
	assert.Equal(t, sig.PublicKey, []byte(key1Pub))
	assert.Equal(t, sig.EntityId, id1)
	assert.NotEmpty(t, sig.Signature)
	assert.Len(t, sig.Signature, 64)
	assert.Equal(t, sig.Signature, signature)
}

// mockIDService implements Service
type mockIDService struct {
	mock.Mock
}

func (srv *mockIDService) LookupIdentityForID(centID []byte) (identity.Identity, error) {
	args := srv.Called(centID)
	id, _ := args.Get(0).(identity.Identity)
	return id, args.Error(1)
}

func (srv *mockIDService) CreateIdentity(centID []byte) (identity.Identity, chan *identity.WatchIdentity, error) {
	args := srv.Called(centID)
	id, _ := args.Get(0).(identity.Identity)
	return id, args.Get(1).(chan *identity.WatchIdentity), args.Error(2)
}

func (srv *mockIDService) CheckIdentityExists(centID []byte) (exists bool, err error) {
	args := srv.Called(centID)
	return args.Bool(0), args.Error(1)
}

func TestValidateSignature_invalid_key(t *testing.T) {
	sig := &coredocumentpb.Signature{EntityId: tools.RandomSlice(identity.CentIdByteLength)}
	srv := &mockIDService{}
	srv.On("LookupIdentityForID", sig.EntityId).Return(nil, fmt.Errorf("failed GetIdentity")).Once()
	identity.SetIdentityService(srv)
	valid, err := ValidateSignature(sig, key1Pub)
	srv.AssertExpectations(t)
	assert.False(t, valid, "should be false")
	assert.NotNil(t, err, "must be not nil")
	assert.Contains(t, err.Error(), "failed GetIdentity")
}

func TestValidateSignature_invalid_sig(t *testing.T) {
	pubKey := key1Pub
	message := key1Pub
	signature := tools.RandomSlice(32)
	valid, err := verifySignature(pubKey, message, signature)
	assert.False(t, valid, "must be false")
	assert.NotNil(t, err, "must be not nil")
	assert.Contains(t, err.Error(), "invalid signature")
}

func TestValidateSignature_success(t *testing.T) {
	pubKey := key1Pub
	message := key1Pub
	valid, err := verifySignature(pubKey, message, signature)
	assert.True(t, valid, "must be true")
	assert.Nil(t, err, "must be nil")
}
