// +build unit

package secp256k1

import (
	"fmt"
	"os"
	"testing"

	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/context/testlogging"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/stretchr/testify/assert"
)

const MaxMsgLen = 32

var ctx = map[string]interface{}{}
var cfg *config.Configuration

func TestMain(m *testing.M) {
	ibootstappers := []bootstrap.TestBootstrapper{
		&config.Bootstrapper{},
		&testlogging.TestLoggingBootstrapper{},
	}
	bootstrap.RunTestBootstrappers(ibootstappers, ctx)
	cfg = ctx[config.BootstrappedConfig].(*config.Configuration)
	result := m.Run()
	os.Exit(result)
}

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

	signature, err := Sign(testMsg, privateKey)
	assert.Nil(t, err)
	correct := VerifySignature(publicKey, testMsg, signature)

	assert.True(t, correct, "sign message didn't work correctly")

}

func TestVerifyFalseMsg(t *testing.T) {

	testMsg := make([]byte, MaxMsgLen)
	copy(testMsg, "test123")

	falseMsg := make([]byte, MaxMsgLen)
	copy(falseMsg, "false")

	publicKey, privateKey := GenerateSigningKeyPair()

	signature, err := Sign(testMsg, privateKey)
	assert.Nil(t, err)
	correct := VerifySignature(publicKey, falseMsg, signature)

	assert.False(t, correct, "false msg verify should be false ")

}

func TestVerifyFalsePublicKey(t *testing.T) {

	testMsg := make([]byte, MaxMsgLen)
	copy(testMsg, "test123")

	_, privateKey := GenerateSigningKeyPair()

	falsePublicKey, _ := GenerateSigningKeyPair()

	signature, err := Sign(testMsg, privateKey)
	assert.Nil(t, err)
	correct := VerifySignature(falsePublicKey, testMsg, signature)

	assert.False(t, correct, "verify of false public key should be false")

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

	assert.True(t, correct, "recovering public key from signature doesn't work correctly")

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

	assert.False(t, correct, "verify signature should be false (false msg)")

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

	assert.False(t, correct, "verify signature should be false (false address)")

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

	assert.False(t, correct, "verify signature should be false (false signature)")

}

func TestSignForEthereum(t *testing.T) {
	privateKey, _ := hexutil.Decode("0xb5fffc3933d93dc956772c69b42c4bc66123631a24e3465956d80b5b604a2d13")
	address := "0xd77c534aed04d7ce34cd425073a033db4fbe6a9d"
	testMsg := "Centrifuge likes Ethereum"

	testMsgBytes := []byte(testMsg)

	//signature should be 0xc158e04b7e22af2380af7b2581c9f89505761d3e517a07fa6bb76889bdb50c604b1517eb4a920053e878478d171ab63c732deb8eb182e3374bcebd046e773a4500
	//verification should work on external services like https://etherscan.io/verifySig
	signature, err := SignEthereum(testMsgBytes, privateKey)
	assert.Nil(t, err)
	sigHex := hexutil.Encode(signature)
	fmt.Println(sigHex)

	correct := VerifySignatureWithAddress(address, sigHex, testMsgBytes)

	assert.Equal(t, correct, true, "generating ethereum signature for msg doesn't work correctly")

}

func TestSignForEthereum32Bytes(t *testing.T) {
	privateKey, _ := hexutil.Decode("0xb5fffc3933d93dc956772c69b42c4bc66123631a24e3465956d80b5b604a2d13")
	address := "0xd77c534aed04d7ce34cd425073a033db4fbe6a9d"

	testMsgBytes := make([]byte, 32)
	copy(testMsgBytes, "Centrifuge likes Ethereum")

	// this signature will not work on external services like etherscan.io because the size of testMsgBytes (32 bytes)
	// is longer than the testMessage in bytes
	signature, err := SignEthereum(testMsgBytes, privateKey)
	assert.Nil(t, err)
	sigHex := hexutil.Encode(signature)

	fmt.Println("address", address)
	fmt.Println("msg:", hexutil.Encode(testMsgBytes))
	fmt.Println("signature:", sigHex)

	correct := VerifySignatureWithAddress(address, sigHex, testMsgBytes)

	assert.Equal(t, correct, true, "generating ethereum signature for msg doesn't work correctly")

}

func TestGetAddress(t *testing.T) {
	//privateKey := "0xde411bde02fdc6998b859241ec6681f29137379cea95c90b1b72540e561a344d"
	publicKey, _ := hexutil.Decode("0x0476464b646617c572f27ee4e0ff7646466bb6cecff1e71f30431cd9ef5f9b163e19bfc5831267fed3818c0b9423386138c2636a0744cf492e12da77e903df592c")
	correctAddress := "0x4ee4a0113f1c833cafcc481e3b3667088607aaa8"

	address := GetAddress(publicKey)
	assert.Equal(t, address, correctAddress, "address is not correctly calculated from public key")
}

func TestGetEthAuthKeyFromConfig(t *testing.T) {
	pub := cfg.Get("keys.ethauth.publicKey")
	pri := cfg.Get("keys.ethauth.privateKey")

	// bad public key path
	cfg.Set("keys.ethauth.publicKey", "bad path")
	pubK, priK, err := GetEthAuthKey(cfg.GetEthAuthKeyPair())
	assert.Error(t, err)
	assert.Nil(t, priK)
	assert.Nil(t, pubK)
	assert.Contains(t, err.Error(), "failed to read public key")
	cfg.Set("keys.ethauth.publicKey", pub)

	// bad private key path
	cfg.Set("keys.ethauth.privateKey", "bad path")
	pubK, priK, err = GetEthAuthKey(cfg.GetEthAuthKeyPair())
	assert.Error(t, err)
	assert.Nil(t, priK)
	assert.Nil(t, pubK)
	assert.Contains(t, err.Error(), "failed to read private key")
	cfg.Set("keys.ethauth.privateKey", pri)

	// success
	pubK, priK, err = GetEthAuthKey(cfg.GetEthAuthKeyPair())
	assert.Nil(t, err)
	assert.NotNil(t, pubK)
	assert.NotNil(t, priK)
}
