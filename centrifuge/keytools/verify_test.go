package keytools

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVerifyMessageED25519(t *testing.T) {

	publicKeyFile := "publicKey"
	privateKeyFile := "privateKey"
	testMsg := "test"

	GenerateSigningKeyPair(publicKeyFile, privateKeyFile, CurveEd25519)
	signature := SignMessage(privateKeyFile, testMsg, CurveEd25519, false)

	os.Remove(publicKeyFile)
	os.Remove(privateKeyFile)

	assert.True(t, len(signature) == 0, "verify ed25519 is not implemented yet and should not work")

}
