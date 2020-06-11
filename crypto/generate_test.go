// +build unit

package crypto

import (
	"os"
	"testing"

	"github.com/centrifuge/go-centrifuge/utils"
	logging "github.com/ipfs/go-log"
	"github.com/stretchr/testify/assert"

	"golang.org/x/crypto/blake2b"
)

const (
	PrivateKeySECP256K1Len = 32
	PublicKeySECP256K1Len  = 65
	PublicKeyED25519Len    = 32
	PrivateKeyED25519Len   = 64
)

var log = logging.Logger("keytools")

func GenerateKeyFilesForTest(t *testing.T, curve string) (publicKey, privateKey []byte) {
	publicFileName := "publicKeyFile"
	privateFileName := "privateKeyFile"
	GenerateSigningKeyPair(publicFileName, privateFileName, curve)

	_, err := os.Stat(publicFileName)
	assert.False(t, err != nil, "public key file not generated")

	_, err = os.Stat(privateFileName)
	assert.False(t, err != nil, "private key file not generated")

	publicKey, err = utils.ReadKeyFromPemFile(publicFileName, utils.PublicKey)
	if err != nil {
		log.Fatal(err)
	}

	privateKey, err = utils.ReadKeyFromPemFile(privateFileName, utils.PrivateKey)
	if err != nil {
		log.Fatal(err)
	}

	os.Remove(publicFileName)
	os.Remove(privateFileName)
	return publicKey, privateKey
}

func TestGenerateSigningKeyPairSECP256K1(t *testing.T) {
	curve := CurveSecp256K1
	publicKey, privateKey := GenerateKeyFilesForTest(t, curve)
	assert.Equal(t, len(publicKey), PublicKeySECP256K1Len, "public key length not correct")
	assert.Equal(t, len(privateKey), PrivateKeySECP256K1Len, "private key length not correct")
}

func TestGenerateSigningKeyPairED25519(t *testing.T) {
	curve := CurveEd25519
	publicKey, privateKey := GenerateKeyFilesForTest(t, curve)
	assert.Equal(t, len(publicKey), PublicKeyED25519Len, "public key length not correct")
	assert.Equal(t, len(privateKey), PrivateKeyED25519Len, "private key length not correct")
}

func TestGenerateHashPair(t *testing.T) {
	pre, hash, err := GenerateHashPair(32)
	assert.NoError(t, err)
	h, err := blake2b.New256(nil)
	assert.NoError(t, err)
	h.Write(pre)
	var expectedHash []byte
	expectedHash = h.Sum(expectedHash)
	assert.Equal(t, expectedHash, hash)
}
