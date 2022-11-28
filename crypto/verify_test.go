//go:build unit

package crypto

import (
	"os"
	"testing"

	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/stretchr/testify/assert"
)

func TestVerifyMessageED25519(t *testing.T) {
	publicKeyFile := "publicKey"
	privateKeyFile := "privateKey"
	testMsg := "test"

	err := GenerateSigningKeyPair(publicKeyFile, privateKeyFile, CurveEd25519)
	assert.NoError(t, err)
	privateKey, err := utils.ReadKeyFromPemFile(privateKeyFile, utils.PrivateKey)
	assert.Nil(t, err)
	signature, err := SignMessage(privateKey, []byte(testMsg), CurveEd25519)
	assert.NoError(t, err)
	assert.Len(t, signature, 64)
	publicKey, err := utils.ReadKeyFromPemFile(publicKeyFile, utils.PublicKey)
	assert.NoError(t, err)
	assert.True(t, VerifyMessage(publicKey, []byte(testMsg), signature, CurveEd25519))

	os.Remove(publicKeyFile)
	os.Remove(privateKeyFile)
}
