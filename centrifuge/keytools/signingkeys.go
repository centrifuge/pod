package keytools

import (
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/keytools/secp256k1"
	logging "github.com/ipfs/go-log"
	"strings"
)

var log = logging.Logger("keytools")

const (
	PUBLIC_KEY  = "PUBLIC KEY"
	PRIVATE_KEY = "PRIVATE KEY"
)

const (
	CURVE_ED25519 string = "ed25519"
	CURVE_SECP256K1 string = "secp256k1"
)

const MAX_MSG_LEN = 32



func SignMessage(privateKeyPath,message, curveType string) ([]byte){

	privateKey, err := readKeyFromPemFile(privateKeyPath, PRIVATE_KEY)

	if(err != nil){
		log.Fatal(err)
	}

	curveType = strings.ToLower(curveType)

	if(len(message) > MAX_MSG_LEN){
		log.Fatal("max message len is 32 bytes current len:", len(message))
	}

	msg := make([]byte, MAX_MSG_LEN)
	copy(msg, message)

	switch (curveType) {

	case CURVE_SECP256K1:
		return secp256k1.Sign(msg,privateKey)
	default:
		return secp256k1.Sign(msg,privateKey)

	}

}

func VerifyMessage(publicKeyPath string,message string,signature []byte,curveType string) (bool) {

	publicKey, err := readKeyFromPemFile(publicKeyPath, PUBLIC_KEY)

	if(err != nil){
		log.Fatal(err)
	}

	msg := make([]byte, MAX_MSG_LEN)
	copy(msg, message)

	signatureBytes := make([]byte, len(signature))
	copy(signatureBytes, signature)

	switch (curveType) {

	case CURVE_SECP256K1:
		return secp256k1.VerifySignature(publicKey,msg,signatureBytes)
	default:
		return secp256k1.VerifySignature(publicKey,msg,signatureBytes)
	}

}

func GenerateSigningKeyPair(publicFileName, privateFileName, curveType string) {

	curveType = strings.ToLower(curveType)

	var publicKey, privateKey []byte

	switch (curveType) {

	case CURVE_SECP256K1:
		publicKey, privateKey = secp256k1.GenerateSigningKeyPair()

	case CURVE_ED25519:
		publicKey, privateKey = GenerateSigningKeyPairED25519()

	default:
		publicKey, privateKey = GenerateSigningKeyPairED25519()

	}
	writeKeyToPemFile(privateFileName, "PRIVATE KEY", privateKey)
	writeKeyToPemFile(publicFileName, "PUBLIC KEY", publicKey)

}

