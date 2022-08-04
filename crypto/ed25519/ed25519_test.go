//go:build unit

package ed25519

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetPublicSigningKey(t *testing.T) {
	fileName := "./build/resources/signingKey.pub.pem"

	key, err := GetPublicSigningKey(fileName)
	assert.NoError(t, err)
	assert.NotNil(t, key)

	key, err = GetPublicSigningKey("")
	assert.NotNil(t, err)
	assert.Nil(t, key)
}

func TestGetPrivateSigningKey(t *testing.T) {
	fileName := "./build/resources/signingKey.key.pem"

	key, err := GetPrivateSigningKey(fileName)
	assert.NoError(t, err)
	assert.NotNil(t, key)

	key, err = GetPrivateSigningKey("")
	assert.NotNil(t, err)
	assert.Nil(t, key)
}

func TestSigningKeyPair(t *testing.T) {
	pubKey, _, err := GenerateSigningKeyPair()
	assert.NoError(t, err)

	var b [32]byte
	copy(b[:], pubKey)

	peerID, err := PublicKeyToP2PKey(b)
	assert.NoError(t, err)
	assert.NotNil(t, peerID)
}
