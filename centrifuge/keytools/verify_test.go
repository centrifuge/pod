package keytools

import (
	"testing"
	"github.com/magiconair/properties/assert"
	"os"
)

func TestVerifyMessageED25519(t *testing.T) {

	publicKeyFile := "publicKey"
	privateKeyFile := "privateKey"
	testMsg := "test"

	GenerateSigningKeyPair(publicKeyFile,privateKeyFile,CURVE_ED25519)
	signature := SignMessage(privateKeyFile,testMsg,CURVE_ED25519)

	os.Remove(publicKeyFile)
	os.Remove(privateKeyFile)

	assert.Equal(t,len(signature) == 0,true,"verify ed25519 is not implemented yet and should not work")

}
