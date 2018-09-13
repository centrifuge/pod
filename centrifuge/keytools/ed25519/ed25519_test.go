// +build unit

package ed25519

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPublicKeyToP2PKey(t *testing.T) {
	expectedPeerId := "QmTQxbwkuZYYDfuzTbxEAReTNCLozyy558vQngVvPMjLYk"
	publicKey := GetPublicSigningKey("../../../example/resources/signingKey.pub.pem")

	var bPk [32]byte
	copy(bPk[:], publicKey)
	peerId, err := PublicKeyToP2PKey(bPk)
	assert.Nil(t, err, "Should not error out")
	assert.Equal(t, expectedPeerId, peerId.Pretty())
}
