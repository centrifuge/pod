package secp256k1

import (
	"fmt"
	"testing"

	"github.com/CentrifugeInc/go-centrifuge/centrifuge/utils"

	"github.com/magiconair/properties/assert"
)

const MaxMsgLen = 32

func TestGenerateSigningKeyPair(t *testing.T) {

	const PublicKeyLen = 65
	const PrivateKeyLen = 32
	publicKey, privateKey := GenerateSigningKeyPair()
	assert.Equal(t, len(publicKey), PublicKeyLen, "secp256k1 public key not correct")
	assert.Equal(t, len(privateKey), PrivateKeyLen, "secp256k1 private key not correct")

}

func TestSigningMsg(t *testing.T) {

	testMsg := make([]byte, MaxMsgLen)
	copy(testMsg, "test123")

	publicKey, privateKey := GenerateSigningKeyPair()

	signature := Sign(testMsg, privateKey)

	correct := VerifySignature(publicKey, testMsg, signature)

	assert.Equal(t, correct, true, "sign message didn't work correctly")

}

func TestVerifyFalseMsg(t *testing.T) {

	testMsg := make([]byte, MaxMsgLen)
	copy(testMsg, "test123")

	falseMsg := make([]byte, MaxMsgLen)
	copy(falseMsg, "false")

	publicKey, privateKey := GenerateSigningKeyPair()

	signature := Sign(testMsg, privateKey)

	correct := VerifySignature(publicKey, falseMsg, signature)

	assert.Equal(t, correct, false, "false msg verify should be false ")

}

func TestVerifyFalsePublicKey(t *testing.T) {

	testMsg := make([]byte, MaxMsgLen)
	copy(testMsg, "test123")

	_, privateKey := GenerateSigningKeyPair()

	falsePublicKey, _ := GenerateSigningKeyPair()

	signature := Sign(testMsg, privateKey)

	correct := VerifySignature(falsePublicKey, testMsg, signature)

	assert.Equal(t, correct, false, "verify of false public key should be false")

}

func TestVerifySignatureWithAddress(t *testing.T) {

	testAddress := "0xd77c534aed04d7ce34cd425073a033db4fbe6a9d"
	//signature generated with an external library (www.myetherwallet.com)
	testSignature := "0x526ea99711a545c745a300e363d277b221d06da2814c521f1b7aa2a3fd0741b85044541da1f985afb51bc4b25a2ab2282721957f694c37a0c68f2fa3220c5cea1c"
	testMsg := "centrifuge"

	correct := VerifySignatureWithAddress(
		testAddress,
		testSignature,
		[]byte(testMsg))

	assert.Equal(t, correct, true, "recovering public key from signature doesn't work correctly")

}

func TestVerifySignatureWithAddressFalseMsg(t *testing.T) {

	testAddress := "0xd77c534aed04d7ce34cd425073a033db4fbe6a9d"
	//signature generated with an external library
	testSignature := "0x526ea99711a545c745a300e363d277b221d06da2814c521f1b7aa2a3fd0741b85044541da1f985afb51bc4b25a2ab2282721957f694c37a0c68f2fa3220c5cea1c"
	falseMsg := "false  msg"

	correct := VerifySignatureWithAddress(
		testAddress,
		testSignature,
		[]byte(falseMsg))

	assert.Equal(t, correct, false, "verify signature should be false (false msg)")

}

func TestVerifySignatureWithFalseAddress(t *testing.T) {

	falseAddress := "0xc8dd3d66e112fae5c88fe6a677be24013e53c33e"
	//signature generated with an external library
	testSignature := "0x526ea99711a545c745a300e363d277b221d06da2814c521f1b7aa2a3fd0741b85044541da1f985afb51bc4b25a2ab2282721957f694c37a0c68f2fa3220c5cea1c"
	testMsg := "centrifuge"

	correct := VerifySignatureWithAddress(
		falseAddress,
		testSignature,
		[]byte(testMsg))

	assert.Equal(t, correct, false, "verify signature should be false (false address)")

}

func TestVerifySignatureWithFalseSignature(t *testing.T) {

	testAddress := "0xd77c534aed04d7ce34cd425073a033db4fbe6a9d"
	//signature generated with an external library
	falseSignature := "0x8efed703a292c278d7de44ab0061144c5bc09a640d168b274b6ad6a7866b7a2542e3e1ae30871d12bf1e882f5b65585a114e9d33615f86e7538f935244071d421b"
	testMsg := "centrifuge"

	correct := VerifySignatureWithAddress(
		testAddress,
		falseSignature,
		[]byte(testMsg))

	assert.Equal(t, correct, false, "verify signature should be false (false signature)")

}

func TestSignForEthereum(t *testing.T) {
	privateKey := "0xb5fffc3933d93dc956772c69b42c4bc66123631a24e3465956d80b5b604a2d13"
	address := "0xd77c534aed04d7ce34cd425073a033db4fbe6a9d"
	testMsg := "centrifuge likes ethereum"

	testMsgBytes := []byte(testMsg)

	//signature should be 0x063e5ae505efd05a028fdf55eeea74434cb9a46efeaa07ffcf1e767620f858981bd9011c1a00e32cda45fb9b9c4f32b5e6e1cdb4bced067942bc4bd78c71c23801
	//verification should work on external services like https://etherscan.io/verifySig
	signature := SignEthereum(testMsgBytes, utils.HexToByteArray(privateKey))

	sigHex := utils.ByteArrayToHex(signature)
	fmt.Println(sigHex)

	correct := VerifySignatureWithAddress(address, sigHex, testMsgBytes)

	assert.Equal(t, correct, true, "generating ethereum signature for msg doesn't work correctly")

}
