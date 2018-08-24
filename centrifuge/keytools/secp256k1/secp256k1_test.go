package secp256k1

import (
	"testing"

	"github.com/magiconair/properties/assert"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/utils"
)

const MAX_MSG_LEN = 32

func TestGenerateSigningKeyPair(t *testing.T) {

	const LEN_PUBLIC_KEY = 65
	const LEN_PRIVATE_KEY = 32
	publicKey, privateKey := GenerateSigningKeyPair()
	assert.Equal(t, len(publicKey), LEN_PUBLIC_KEY, "secp256k1 public key not correct")
	assert.Equal(t, len(privateKey), LEN_PRIVATE_KEY, "secp256k1 private key not correct")

}

func TestSigningMsg(t *testing.T) {

	testMsg := make([]byte, MAX_MSG_LEN)
	copy(testMsg, "test123")

	publicKey, privateKey := GenerateSigningKeyPair()

	signature := Sign(testMsg, privateKey)

	correct := VerifySignature(publicKey, testMsg, signature)

	assert.Equal(t, correct, true, "sign message didn't work correctly")

}

func TestVerifyFalseMsg(t *testing.T) {

	testMsg := make([]byte, MAX_MSG_LEN)
	copy(testMsg, "test123")

	falseMsg := make([]byte, MAX_MSG_LEN)
	copy(falseMsg, "false")

	publicKey, privateKey := GenerateSigningKeyPair()

	signature := Sign(testMsg, privateKey)

	correct := VerifySignature(publicKey, falseMsg, signature)

	assert.Equal(t, correct, false, "false msg verify should be false ")

}

func TestVerifyFalsePublicKey(t *testing.T) {

	testMsg := make([]byte, MAX_MSG_LEN)
	copy(testMsg, "test123")

	_, privateKey := GenerateSigningKeyPair()

	falsePublicKey, _ := GenerateSigningKeyPair()

	signature := Sign(testMsg, privateKey)

	correct := VerifySignature(falsePublicKey, testMsg, signature)

	assert.Equal(t, correct, false, "verify of false public key should be false")

}

func TestVerifySignatureWithAddress(t *testing.T) {


	testAddress := "0xd77c534aed04d7ce34cd425073a033db4fbe6a9d"
	testMsg := make([]byte, MAX_MSG_LEN)
	copy(testMsg, "centrifuge")
	testSignature := "0x526ea99711a545c745a300e363d277b221d06da2814c521f1b7aa2a3fd0741b85044541da1f985afb51bc4b25a2ab2282721957f694c37a0c68f2fa3220c5cea1c"

	testSignatureBytes, err := utils.HexToByteArray(testSignature)

	if(err != nil){
		log.Fatal(err)
	}

	correct := VerifySignatureWithAddress(testAddress,testMsg,testSignatureBytes)

	assert.Equal(t,correct,true,"address from signature not correctly calculated")

}


