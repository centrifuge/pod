// +build unit

package signatures

import (
	"fmt"
	"testing"

	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/config"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/identity"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/testingutils"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/testingutils/commons"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/tools"
	"github.com/stretchr/testify/assert"
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

	sig := Sign(&idConfig, coreDoc.SigningRoot)
	assert.NotNil(t, sig)
	assert.Equal(t, sig.PublicKey, []byte(key1Pub))
	assert.Equal(t, sig.EntityId, id1)
	assert.NotEmpty(t, sig.Signature)
	assert.Len(t, sig.Signature, 64)
	assert.Equal(t, sig.Signature, signature)
	assert.NotNil(t, sig.Timestamp, "must be non nil")
}

func TestValidateSignature_invalid_key(t *testing.T) {
	sig := &coredocumentpb.Signature{EntityId: tools.RandomSlice(identity.CentIDByteLength)}
	srv := &testingcommons.MockIDService{}
	centId, _ := identity.NewCentID(sig.EntityId)
	srv.On("LookupIdentityForID", centId).Return(nil, fmt.Errorf("failed GetIdentity")).Once()
	identity.IDService = srv
	err := ValidateSignature(sig, key1Pub)
	srv.AssertExpectations(t)
	assert.NotNil(t, err, "must be not nil")
	assert.Contains(t, err.Error(), "failed GetIdentity")
}

func TestValidateSignature_invalid_sig(t *testing.T) {
	pubKey := key1Pub
	message := key1Pub
	signature := tools.RandomSlice(32)
	err := verifySignature(pubKey, message, signature)
	assert.NotNil(t, err, "must be not nil")
	assert.Contains(t, err.Error(), "invalid signature")
}

func TestValidateSignature_success(t *testing.T) {
	pubKey := key1Pub
	message := key1Pub
	err := verifySignature(pubKey, message, signature)
	assert.Nil(t, err, "must be nil")
}

func TestValidateCentrifugeId(t *testing.T) {

	randomBytes := tools.RandomSlice(identity.CentIDByteLength)

	centrifugeId, err := identity.NewCentID(randomBytes)

	assert.Nil(t, err, "centrifugeId not initialized correctly ")

	sig := &coredocumentpb.Signature{EntityId: randomBytes}

	err = ValidateCentrifugeID(sig, centrifugeId)
	assert.Nil(t, err, "Validate centrifuge id didn't work correctly")

	randomBytes = tools.RandomSlice(identity.CentIDByteLength)
	centrifugeId, err = identity.NewCentID(randomBytes)

	err = ValidateCentrifugeID(sig, centrifugeId)
	assert.NotNil(t, err, "Validate centrifuge id didn't work correctly")

}
