package keytools

import (
	"testing"
	"github.com/magiconair/properties/assert"
	"os"
)

func TestSignMessage(t *testing.T) {

	publicKeyFile := "publicKey"
	privateKeyFile := "privateKey"
	testMsg := "test"

	GenerateSigningKeyPair(publicKeyFile,privateKeyFile,CURVE_SECP256K1)
	signature := SignMessage(privateKeyFile,"test",CURVE_SECP256K1)

	correct := VerifyMessage(publicKeyFile,testMsg,signature,CURVE_SECP256K1)

	os.Remove(publicKeyFile)
	os.Remove(privateKeyFile)

	assert.Equal(t,correct,true,"signature or verification didn't work correctly")

}
