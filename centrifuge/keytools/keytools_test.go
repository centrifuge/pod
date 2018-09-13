// +build unit

package keytools

import (
	"os"
	"testing"

	"github.com/CentrifugeInc/go-centrifuge/centrifuge/tools"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/utils"
	"github.com/stretchr/testify/assert"
)

const (
	PrivateKeySECP256K1Len = 32
	PublicKeySECP256K1Len  = 65
	PublicKeyED25519Len    = 32
	PrivateKeyED25519Len   = 64
)

func TestSaveKeyPair(t *testing.T) {
	publicFileName := "publicKeyFile"
	privateFileName := "privateKeyFile"

	pub, priv := GenerateSigningKeyPair(CurveEd25519)
	err := SaveKeyPair(publicFileName, privateFileName, pub, priv)
	assert.Nil(t, err, "must be nil")

	_, err = os.Stat(publicFileName)
	assert.Nil(t, err, "public key file not generated")

	_, err = os.Stat(privateFileName)
	assert.Nil(t, err, "private key file not generated")

	publicKey, err := utils.ReadKeyFromPemFile(publicFileName, PublicKey)
	assert.Nil(t, err, "must be nil")
	assert.NotNil(t, publicKey, "key must not be nil")
	assert.Equal(t, publicKey, pub, "public keys must match")

	privateKey, err := utils.ReadKeyFromPemFile(privateFileName, PrivateKey)
	assert.Nil(t, err, "must be nil")
	assert.NotNil(t, privateKey, "key must not be nil")
	assert.Equal(t, privateKey, priv, "private keys must match")

	os.Remove(publicFileName)
	os.Remove(privateFileName)
}

func TestGenerateSigningKeyPairSECP256K1(t *testing.T) {
	curve := CurveSecp256K1
	publicKey, privateKey := GenerateSigningKeyPair(curve)
	assert.Equal(t, len(publicKey), PublicKeySECP256K1Len, "public key length not correct")
	assert.Equal(t, len(privateKey), PrivateKeySECP256K1Len, "private key length not correct")
}

func TestGenerateSigningKeyPairED25519(t *testing.T) {
	curve := CurveEd25519
	publicKey, privateKey := GenerateSigningKeyPair(curve)
	assert.Equal(t, len(publicKey), PublicKeyED25519Len, "public key length not correct")
	assert.Equal(t, len(privateKey), PrivateKeyED25519Len, "private key length not correct")
}

func TestGenerateSigningKeyPairUnknown(t *testing.T) {
	curve := "SomeCurve"
	publicKey, privateKey := GenerateSigningKeyPair(curve)
	assert.Equal(t, len(publicKey), PublicKeyED25519Len, "public key length not correct")
	assert.Equal(t, len(privateKey), PrivateKeyED25519Len, "private key length not correct")
}

func TestSignAndVerifyMessage_secp256K1(t *testing.T) {
	testMsg := []byte("test")
	publicKey, privateKey := GenerateSigningKeyPair(CurveSecp256K1)
	signature, err := SignMessage(privateKey, testMsg, CurveSecp256K1, false)
	assert.Nil(t, err)
	correct := VerifyMessage(publicKey, testMsg, signature, CurveSecp256K1, false)
	assert.True(t, correct, "signature or verification didn't work correctly")
}

func TestSignAndVerifyMessage_secp256k1_ethereum(t *testing.T) {
	testMsg := []byte("Centrifuge likes Ethereum")
	publicKey, privateKey := GenerateSigningKeyPair(CurveSecp256K1)
	signature, err := SignMessage(privateKey, testMsg, CurveSecp256K1, true)
	assert.Nil(t, err)
	correct := VerifyMessage(publicKey, testMsg, signature, CurveSecp256K1, true)
	assert.True(t, correct, "signature or verification didn't work correctly")
}

func TestSignMessage_ed25519(t *testing.T) {
	testMsg := []byte("test")
	_, privateKey := GenerateSigningKeyPair(CurveEd25519)
	signature, err := SignMessage(privateKey, testMsg, CurveEd25519, false)
	assert.Nil(t, signature, "must be nil")
	assert.Error(t, err, "must return non nil error")
	assert.Contains(t, err.Error(), "curve ed25519 not supported")
}

func TestVerifyMessageED25519(t *testing.T) {
	testMsg := "test"
	public, _ := GenerateSigningKeyPair(CurveEd25519)
	valid := VerifyMessage(public, []byte(testMsg), tools.RandomSlice(32), CurveEd25519, false)
	assert.False(t, valid, "should be false")
}
