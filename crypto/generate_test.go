//go:build unit

package crypto

import (
	"os"
	"testing"

	"github.com/centrifuge/pod/utils"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/blake2b"
)

const (
	PublicKeyED25519Len  = 32
	PrivateKeyED25519Len = 64
)

func GenerateKeyFilesForTest(t *testing.T, curve CurveType) (publicKey, privateKey []byte) {
	publicFileName := "publicKeyFile"
	privateFileName := "privateKeyFile"
	err := GenerateSigningKeyPair(publicFileName, privateFileName, curve)
	assert.NoError(t, err)

	_, err = os.Stat(publicFileName)
	assert.False(t, err != nil, "public key file not generated")

	_, err = os.Stat(privateFileName)
	assert.False(t, err != nil, "private key file not generated")

	publicKey, err = utils.ReadKeyFromPemFile(publicFileName, utils.PublicKey)
	if err != nil {
		t.Fatal(err)
	}

	privateKey, err = utils.ReadKeyFromPemFile(privateFileName, utils.PrivateKey)
	if err != nil {
		t.Fatal(err)
	}

	_ = os.Remove(publicFileName)
	_ = os.Remove(privateFileName)
	return publicKey, privateKey
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
