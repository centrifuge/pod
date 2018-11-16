// +build unit

package keytools

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
	assert.NotNil(t, err)
	os.Remove(publicKeyFile)
	os.Remove(privateKeyFile)
	assert.True(t, len(signature) == 0, "verify ed25519 is not implemented yet and should not work")
}
