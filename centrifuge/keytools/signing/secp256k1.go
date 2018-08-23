package signing

import (
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/elliptic"
	"github.com/ethereum/go-ethereum/crypto/secp256k1"
	logging "github.com/ipfs/go-log"
)

var log = logging.Logger("signing")

const LEN_SIGNATURE = 64 //64 byte [R || S] format

func GenerateSigningKeyPairSECP256K1() (publicKey, privateKey []byte) {

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

func SignSECP256K1(message []byte, privateKey []byte) (signature []byte){
	signature, err := secp256k1.Sign(message,privateKey)

	if err != nil {
		log.Fatal(err)
	}
	return signature

}

func VerifySignatureSECP256K1(publicKey,message,signature []byte)(bool){
    if(len(signature) == LEN_SIGNATURE+1){
    	signature = signature[0:LEN_SIGNATURE]
	}
	// the signature should have the 64 byte [R || S] format
	return secp256k1.VerifySignature(publicKey,message,signature)

}

