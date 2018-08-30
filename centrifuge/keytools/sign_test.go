package keytools

import (
	"fmt"
	"os"
	"testing"

	"github.com/CentrifugeInc/go-centrifuge/centrifuge/keytools/io"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/keytools/secp256k1"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/utils"

	"github.com/magiconair/properties/assert"
)

func TestSignMessage(t *testing.T) {

	publicKeyFile := "publicKey"
	privateKeyFile := "privateKey"
	testMsg := "test"

	GenerateSigningKeyPair(publicKeyFile, privateKeyFile, CurveSecp256K1)
	signature := SignMessage(privateKeyFile, "test", CurveSecp256K1, false)

	correct := VerifyMessage(publicKeyFile, testMsg, signature, CurveSecp256K1, false)

	os.Remove(publicKeyFile)
	os.Remove(privateKeyFile)

	assert.Equal(t, correct, true, "signature or verification didn't work correctly")

}

func TestSignAndVerifyMessageEthereum(t *testing.T) {

	publicKeyFile := "publicKey"
	privateKeyFile := "privateKey"
	testMsg := "Centrifuge likes Ethereum"

	GenerateSigningKeyPair(publicKeyFile, privateKeyFile, CurveSecp256K1)
	signature := SignMessage(privateKeyFile, testMsg, CurveSecp256K1, true)

	privateKey, _ := io.ReadKeyFromPemFile(privateKeyFile, PrivateKey)
	publicKey, _ := io.ReadKeyFromPemFile(publicKeyFile, PublicKey)
	address := secp256k1.GetAddress(publicKey)

	fmt.Println("privateKey: ", utils.ByteArrayToHex(privateKey))
	fmt.Println("publicKey: ", utils.ByteArrayToHex(publicKey))
	fmt.Println("address:", address)
	fmt.Println("msg:", testMsg)
	fmt.Println("msg in hex:",utils.ByteArrayToHex([]byte(testMsg)))
	fmt.Println("hash of msg: ", utils.ByteArrayToHex(secp256k1.SignHash([]byte(testMsg))))
	fmt.Println("signature:", utils.ByteArrayToHex(signature))
	fmt.Println("Generated Signature can also be verified at https://etherscan.io/verifySig")

	correct := VerifyMessage(publicKeyFile, testMsg, signature, CurveSecp256K1, true)

	os.Remove(publicKeyFile)
	os.Remove(privateKeyFile)

	assert.Equal(t, correct, true, "signature or verification didn't work correctly")

}
