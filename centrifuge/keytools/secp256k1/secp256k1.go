package secp256k1

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"

	"github.com/ethereum/go-ethereum/crypto/secp256k1"
	"github.com/ethereum/go-ethereum/crypto"
	logging "github.com/ipfs/go-log"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/utils"
	"fmt"
)

var log = logging.Logger("signing")

const LEN_SIGNATURE_R_S_FORMAT = 64 //64 byte [R || S] format

const LEN_OF_ADDRESS = 20

func GenerateSigningKeyPair() (publicKey, privateKey []byte) {

	log.Debug("generate secp256k1 keys")
	key, err := ecdsa.GenerateKey(secp256k1.S256(), rand.Reader)
	if err != nil {
		log.Fatal(err)
	}
	publicKey = elliptic.Marshal(secp256k1.S256(), key.X, key.Y)

	privateKey = make([]byte, 32)
	blob := key.D.Bytes()
	copy(privateKey[32-len(blob):], blob)

	return publicKey, privateKey
}

func Sign(message []byte, privateKey []byte) (signature []byte) {
	signature, err := secp256k1.Sign(message, privateKey)

	if err != nil {
		log.Fatal(err)
	}
	return signature

}


func VerifySignatureWithAddress(address string, message, signature []byte) bool {

	//how addresses are generated in ethereum
	//https://ethereum.stackexchange.com/questions/3542/how-are-ethereum-addresses-generated

	publicKey, err := secp256k1.RecoverPubkey(message, signature)

	if (err != nil) {
		log.Fatal(err)
	}
	hash := crypto.Keccak256(publicKey)

	addressFromSignature := utils.ByteArrayToHex(hash[len(hash)-LEN_OF_ADDRESS:])

	fmt.Println(addressFromSignature)

	return addressFromSignature == address

}


func VerifySignature(publicKey, message, signature []byte) bool {
	if len(signature) == LEN_SIGNATURE_R_S_FORMAT+1 {
		// signature in [R || S || V] format is 65 bytes
		//https://bitcoin.stackexchange.com/questions/38351/ecdsa-v-r-s-what-is-v

		signature = signature[0:LEN_SIGNATURE_R_S_FORMAT]
	}
	// the signature should have the 64 byte [R || S] format
	return secp256k1.VerifySignature(publicKey, message, signature)

}
