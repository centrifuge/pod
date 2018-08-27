package keytools

import (
	"os"
	"testing"

	"github.com/magiconair/properties/assert"
)

func TestSignMessage(t *testing.T) {

	publicKeyFile := "publicKey"
	privateKeyFile := "privateKey"
	testMsg := "test"

	GenerateSigningKeyPair(publicKeyFile, privateKeyFile, CurveSecp256K1)
	signature := SignMessage(privateKeyFile, "test", CurveSecp256K1)

	correct := VerifyMessage(publicKeyFile, testMsg, signature, CurveSecp256K1)

	os.Remove(publicKeyFile)
	os.Remove(privateKeyFile)

	assert.Equal(t, correct, true, "signature or verification didn't work correctly")

}
