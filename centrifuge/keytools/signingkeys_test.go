// +build unit

package keytools

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPublicKeyToP2PKey(t *testing.T) {
	expectedPeerId := "QmTQxbwkuZYYDfuzTbxEAReTNCLozyy558vQngVvPMjLYk"
	publicKey := GetPublicSigningKey("../../example/resources/signingKey.pub")

	var bPk [32]byte
	copy(bPk[:], publicKey)
	peerId, err := PublicKeyToP2PKey(bPk)
	assert.Nil(t, err, "Should not error out")
	assert.Equal(t, expectedPeerId, peerId.Pretty())

}
