package secp256k1

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"

	"github.com/ethereum/go-ethereum/crypto/secp256k1"
	"github.com/ethereum/go-ethereum/crypto"
	logging "github.com/ipfs/go-log"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/common"
	"fmt"
)

var log = logging.Logger("signing")

const LEN_SIGNATURE_R_S_FORMAT = 64 //64 byte [R || S] format
const LEN_SIGNATURE_R_S_V_FORMAT = 65 //65 byte [R || S || V] format
const SIGNATURE_V_POSITION  = 64

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


func VerifySignatureWithAddress(address, sigHex string, msg []byte) bool {
	fromAddr := common.HexToAddress(address)

	sig := hexutil.MustDecode(sigHex)

	if(len(sig) != LEN_SIGNATURE_R_S_V_FORMAT){
		log.Fatal("signature must be 65 bytes long")
		return false
	}

	// see implementation in go-ethereum for further details
	// https://github.com/ethereum/go-ethereum/blob/55599ee95d4151a2502465e0afc7c47bd1acba77/internal/ethapi/api.go#L442
	if sig[SIGNATURE_V_POSITION] != 27 && sig[SIGNATURE_V_POSITION] != 28 {
		log.Fatal("V value in signature has to be 27 or 28")
		return false
	}
	sig[SIGNATURE_V_POSITION] -= 27 // change V value to 0 or 1

	pubKey, err := crypto.SigToPub(signHash(msg), sig)
	if err != nil {

		return false
	}

	recoveredAddr := crypto.PubkeyToAddress(*pubKey)

	return fromAddr == recoveredAddr
}

// For more information
// https://github.com/ethereum/go-ethereum/blob/55599ee95d4151a2502465e0afc7c47bd1acba77/internal/ethapi/api.go#L404
// The hash is calculated as
//   keccak256("\x19Ethereum Signed Message:\n"${message length}${message}).
//
func signHash(data []byte) []byte {
	msg := fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(data), data)
	fmt.Println(msg)
	return crypto.Keccak256([]byte(msg))
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
