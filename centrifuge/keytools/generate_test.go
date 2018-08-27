package keytools

import (
	"testing"
	"os"
	"github.com/magiconair/properties/assert"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/keytools/io"
)

const (
	PRIVATE_KEY_SECP256K1_LEN = 32
	PUBLIC_KEY_SECP256K1_LEN = 65
	PUBLIC_KEY_ED25519_LEN = 32
	PRIVATE_KEY_ED25519_LEN = 64
)

func GenerateKeyFilesForTest(t *testing.T,curve string) (publicKey, privateKey []byte){

	publicFileName := "publicKeyFile"
	privateFileName := "privateKeyFile"

	GenerateSigningKeyPair(publicFileName,privateFileName,curve)

	_, err := os.Stat(publicFileName)

	assert.Equal(t,err != nil,false, "public key file not generated")

	_, err = os.Stat(privateFileName)

	assert.Equal(t,err != nil,false, "private key file not generated")

	publicKey, err = io.ReadKeyFromPemFile(publicFileName,PUBLIC_KEY)

	if err != nil {
		log.Fatal(err)
	}

	privateKey, err = io.ReadKeyFromPemFile(privateFileName,PRIVATE_KEY)

	if err != nil {
		log.Fatal(err)
	}

	os.Remove(publicFileName)
	os.Remove(privateFileName)

	return

}

func TestGenerateSigningKeyPairSECP256K1(t *testing.T) {

	curve := CURVE_SECP256K1
	publicKey,privateKey := GenerateKeyFilesForTest(t,curve)

	assert.Equal(t,len(publicKey),PUBLIC_KEY_SECP256K1_LEN,"public key length not correct")
	assert.Equal(t,len(privateKey),PRIVATE_KEY_SECP256K1_LEN,"private key length not correct")

}

func TestGenerateSigningKeyPairED25519(t *testing.T) {

	curve := CURVE_ED25519
	publicKey,privateKey := GenerateKeyFilesForTest(t,curve)

	assert.Equal(t,len(publicKey),PUBLIC_KEY_ED25519_LEN,"public key length not correct")
	assert.Equal(t,len(privateKey),PRIVATE_KEY_ED25519_LEN,"private key length not correct")
}
