// +build unit

package crypto

import (
	"fmt"
	"os"
	"testing"

	"github.com/centrifuge/go-centrifuge/crypto/secp256k1"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/stretchr/testify/assert"
)

var (
	key1Pub   = []byte{230, 49, 10, 12, 200, 149, 43, 184, 145, 87, 163, 252, 114, 31, 91, 163, 24, 237, 36, 51, 165, 8, 34, 104, 97, 49, 114, 85, 255, 15, 195, 199}
	key1      = []byte{102, 109, 71, 239, 130, 229, 128, 189, 37, 96, 223, 5, 189, 91, 210, 47, 89, 4, 165, 6, 188, 53, 49, 250, 109, 151, 234, 139, 57, 205, 231, 253, 230, 49, 10, 12, 200, 149, 43, 184, 145, 87, 163, 252, 114, 31, 91, 163, 24, 237, 36, 51, 165, 8, 34, 104, 97, 49, 114, 85, 255, 15, 195, 199}
	id1       = []byte{1, 1, 1, 1, 1, 1}
	signature = []byte{0x4e, 0x3d, 0x90, 0x5f, 0x25, 0xc7, 0x90, 0x63, 0x7e, 0x6c, 0xd0, 0xe6, 0xc7, 0xbd, 0xe6, 0x81, 0x3b, 0xd0, 0x5b, 0x94, 0x76, 0x86, 0x4e, 0xcb, 0xb9, 0x36, 0x48, 0x44, 0x4b, 0x98, 0xd2, 0x4b, 0x6a, 0x65, 0x22, 0x92, 0x1c, 0x8a, 0xdb, 0xfe, 0xb7, 0x6f, 0xfe, 0x34, 0x52, 0xa3, 0x49, 0xe4, 0xda, 0xdc, 0x5d, 0x1b, 0x0, 0x79, 0x54, 0x60, 0x29, 0x22, 0x94, 0xb, 0x3c, 0x90, 0x3c, 0x3}
)

func TestSign(t *testing.T) {
	sig := Sign(id1, key1, key1Pub, key1Pub)
	assert.NotNil(t, sig)
	assert.Equal(t, sig.PublicKey, []byte(key1Pub))
	assert.Equal(t, sig.EntityId, id1)
	assert.NotEmpty(t, sig.Signature)
	assert.Len(t, sig.Signature, 64)
	assert.Equal(t, sig.Signature, signature)
	assert.NotNil(t, sig.Timestamp, "must be non nil")
}

func TestValidateSignature_invalid_sig(t *testing.T) {
	pubKey := key1Pub
	message := key1Pub
	signature := utils.RandomSlice(32)
	err := VerifySignature(pubKey, message, signature)
	assert.NotNil(t, err, "must be not nil")
	assert.Contains(t, err.Error(), "invalid signature")
}

func TestValidateSignature_success(t *testing.T) {
	pubKey := key1Pub
	message := key1Pub
	err := VerifySignature(pubKey, message, signature)
	assert.Nil(t, err, "must be nil")
}

func TestSignMessage(t *testing.T) {

	publicKeyFile := "publicKey"
	privateKeyFile := "privateKey"
	testMsg := []byte("test")

	GenerateSigningKeyPair(publicKeyFile, privateKeyFile, CurveSecp256K1)
	privateKey, err := utils.ReadKeyFromPemFile(privateKeyFile, utils.PrivateKey)
	assert.Nil(t, err)
	publicKey, err := utils.ReadKeyFromPemFile(publicKeyFile, utils.PublicKey)
	assert.Nil(t, err)
	signature, err := SignMessage(privateKey, testMsg, CurveSecp256K1, false)
	assert.Nil(t, err)
	correct := VerifyMessage(publicKey, testMsg, signature, CurveSecp256K1, false)

	os.Remove(publicKeyFile)
	os.Remove(privateKeyFile)

	assert.True(t, correct, "signature or verification didn't work correctly")
}

func TestSignAndVerifyMessageEthereum(t *testing.T) {

	publicKeyFile := "publicKey"
	privateKeyFile := "privateKey"
	testMsg := []byte("Centrifuge likes Ethereum")

	GenerateSigningKeyPair(publicKeyFile, privateKeyFile, CurveSecp256K1)
	privateKey, err := utils.ReadKeyFromPemFile(privateKeyFile, utils.PrivateKey)
	assert.Nil(t, err)
	signature, err := SignMessage(privateKey, testMsg, CurveSecp256K1, true)
	assert.Nil(t, err)

	publicKey, _ := utils.ReadKeyFromPemFile(publicKeyFile, utils.PublicKey)
	address := secp256k1.GetAddress(publicKey)

	fmt.Println("privateKey: ", hexutil.Encode(privateKey))
	fmt.Println("publicKey: ", hexutil.Encode(publicKey))
	fmt.Println("address:", address)
	fmt.Println("msg:", string(testMsg[:]))
	fmt.Println("msg in hex:", hexutil.Encode(testMsg))
	fmt.Println("hash of msg: ", hexutil.Encode(secp256k1.SignHash(testMsg)))
	fmt.Println("signature:", hexutil.Encode(signature))
	fmt.Println("Generated Signature can also be verified at https://etherscan.io/verifySig")

	correct := VerifyMessage(publicKey, testMsg, signature, CurveSecp256K1, true)

	os.Remove(publicKeyFile)
	os.Remove(privateKeyFile)

	assert.True(t, correct, "signature or verification didn't work correctly")
}
