// +build unit

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

	GenerateSigningKeyPair(publicKeyFile, privateKeyFile, CurveEd25519)
	privateKey, err := utils.ReadKeyFromPemFile(privateKeyFile, utils.PrivateKey)
	assert.Nil(t, err)
	signature, err := SignMessage(privateKey, []byte(testMsg), CurveEd25519, false)
	assert.NoError(t, err)
	os.Remove(publicKeyFile)
	os.Remove(privateKeyFile)
	assert.Len(t, signature, 64)
}
