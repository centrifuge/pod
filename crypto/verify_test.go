// +build unit

package crypto

import (
	"os"
	"testing"

	"github.com/centrifuge/go-centrifuge/crypto/secp256k1"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
)

func TestVerifyMessageED25519(t *testing.T) {
	publicKeyFile := "publicKey"
	privateKeyFile := "privateKey"
	testMsg := "test"

	err := GenerateSigningKeyPair(publicKeyFile, privateKeyFile, CurveEd25519)
	assert.NoError(t, err)
	privateKey, err := utils.ReadKeyFromPemFile(privateKeyFile, utils.PrivateKey)
	assert.Nil(t, err)
	signature, err := SignMessage(privateKey, []byte(testMsg), CurveEd25519)
	assert.NoError(t, err)
	assert.Len(t, signature, 64)
	publicKey, err := utils.ReadKeyFromPemFile(publicKeyFile, utils.PublicKey)
	assert.NoError(t, err)
	assert.True(t, VerifyMessage(publicKey, []byte(testMsg), signature, CurveEd25519))

	os.Remove(publicKeyFile)
	os.Remove(privateKeyFile)
}

func TestVerifyMessageSecp256k1(t *testing.T) {
	publicKeyFile := "publicKey"
	privateKeyFile := "privateKey"
	testMsg := "test"

	err := GenerateSigningKeyPair(publicKeyFile, privateKeyFile, CurveSecp256K1)
	assert.NoError(t, err)
	privateKey, err := utils.ReadKeyFromPemFile(privateKeyFile, utils.PrivateKey)
	assert.Nil(t, err)
	signature, err := SignMessage(privateKey, []byte(testMsg), CurveSecp256K1)
	assert.NoError(t, err)
	assert.Len(t, signature, 65)
	publicKey, err := utils.ReadKeyFromPemFile(publicKeyFile, utils.PublicKey)
	assert.NoError(t, err)
	pk32 := utils.AddressTo32Bytes(common.HexToAddress(secp256k1.GetAddress(publicKey)))
	assert.True(t, VerifyMessage(pk32[:], []byte(testMsg), signature, CurveSecp256K1))

	os.Remove(publicKeyFile)
	os.Remove(privateKeyFile)
}
