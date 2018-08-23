package signing

import (
	"testing"
	"github.com/magiconair/properties/assert"
)

const MAX_MSG_LEN = 32

func TestGenerateSigningKeyPairSECP256K1(t *testing.T) {

	const LEN_PUBLIC_KEY = 65
	const LEN_PRIVATE_KEY = 32
	publicKey, privateKey := GenerateSigningKeyPairSECP256K1()
	assert.Equal(t,len(publicKey),LEN_PUBLIC_KEY,"secp256k1 public key not correct")
	assert.Equal(t,len(privateKey),LEN_PRIVATE_KEY,"secp256k1 private key not correct")

}

func TestSigningMsgSECP256K1(t *testing.T) {

	testMsg := make([]byte, MAX_MSG_LEN)
	copy(testMsg, "test123")

	publicKey, privateKey := GenerateSigningKeyPairSECP256K1()

	signature := SignSECP256K1(testMsg,privateKey)

	correct := VerifySignatureSECP256K1(publicKey,testMsg,signature)

	assert.Equal(t,correct,true,"sign message didn't work correctly")

}

func TestVerifyFalseMsgSECP256K1(t *testing.T) {

	testMsg := make([]byte, MAX_MSG_LEN)
	copy(testMsg, "test123")

	falseMsg := make([]byte, MAX_MSG_LEN)
	copy(falseMsg, "false")

	publicKey, privateKey := GenerateSigningKeyPairSECP256K1()

	signature := SignSECP256K1(testMsg,privateKey)

	correct := VerifySignatureSECP256K1(publicKey,falseMsg,signature)

	assert.Equal(t,correct,false,"false msg verify should be false ")

}

func TestVerifyFalsePublicKeySECP256K1(t *testing.T) {

	testMsg := make([]byte, MAX_MSG_LEN)
	copy(testMsg, "test123")

	_, privateKey := GenerateSigningKeyPairSECP256K1()

	falsePublicKey, _ := GenerateSigningKeyPairSECP256K1()

	signature := SignSECP256K1(testMsg,privateKey)

	correct := VerifySignatureSECP256K1(falsePublicKey,testMsg,signature)

	assert.Equal(t,correct,false,"verify of false public key should be false")

}

