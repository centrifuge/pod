package secp256k1

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"fmt"

	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/crypto/secp256k1"
	logging "github.com/ipfs/go-log"
)

var log = logging.Logger("signing")

const (
	signatureRSFormatLen  = 64 //64 byte [R || S] format
	signatureRSVFormatLen = 65 //65 byte [R || S || V] format
	signatureVPosition    = 64
	privateKeyLen         = 32
)

// GenerateSigningKeyPair generates secp2562k1 based keys.
func GenerateSigningKeyPair() (publicKey, privateKey []byte, err error) {
	log.Debug("generate secp256k1 keys")
	key, err := ecdsa.GenerateKey(secp256k1.S256(), rand.Reader)
	if err != nil {
		return []byte{}, []byte{}, nil
	}
	publicKey = elliptic.Marshal(secp256k1.S256(), key.X, key.Y)

	privateKey = make([]byte, privateKeyLen)
	blob := key.D.Bytes()
	copy(privateKey[privateKeyLen-len(blob):], blob)

	return publicKey, privateKey, nil
}

// Sign signs the message using private key
// We do hash the message since it not recommended to use the message as is.
func Sign(message []byte, privateKey []byte) (signature []byte, err error) {
	return secp256k1.Sign(Hash(message), privateKey)
}

// SignEthereum converts message to ethereum specific format and signs it.
func SignEthereum(message []byte, privateKey []byte) (signature []byte, err error) {
	return secp256k1.Sign(HashWithEthPrefix(message), privateKey)
}

// GetAddress returns the hex of first 20 bytes of the Keccak256 has of public keuy
func GetAddress(publicKey []byte) string {
	hash := crypto.Keccak256(publicKey[1:])
	address := hash[12:] //address is the last 20 bytes of the hash len(hash) = 20
	return hexutil.Encode(address)
}

// VerifySignatureWithAddress verifies the signature using address provided
func VerifySignatureWithAddress(address, sigHex string, msg []byte) bool {
	fromAddr := common.HexToAddress(address)
	sig, err := hexutil.Decode(sigHex)
	if err != nil {
		log.Error(err.Error())
		return false
	}

	if len(sig) != signatureRSVFormatLen {
		log.Error("signature must be 65 bytes long")
		return false
	}

	// see implementation in go-ethereum for further details
	// https://github.com/ethereum/go-ethereum/blob/55599ee95d4151a2502465e0afc7c47bd1acba77/internal/ethapi/api.go#L442
	if sig[signatureVPosition] != 0 && sig[signatureVPosition] != 1 {
		if sig[signatureVPosition] != 27 && sig[signatureVPosition] != 28 {
			log.Error("V value in signature has to be 27 or 28")
			return false
		}
		sig[signatureVPosition] -= 27 // change V value to 0 or 1
	}

	pubKey, err := crypto.SigToPub(HashWithEthPrefix(msg), sig)
	if err != nil {
		return false
	}

	recoveredAddr := crypto.PubkeyToAddress(*pubKey)
	return fromAddr == recoveredAddr
}

// HashWithEthPrefix returns the hash of the data.
// The hash is calculated as
// keccak256("\x19Ethereum Signed Message:\n"${message length}${message}).
// for further details see
// https://github.com/ethereum/go-ethereum/blob/55599ee95d4151a2502465e0afc7c47bd1acba77/internal/ethapi/api.go#L404
func HashWithEthPrefix(data []byte) []byte {
	msg := fmt.Sprintf("\x19Ethereum Signed Message:\n%d", len(data))
	return crypto.Keccak256(append([]byte(msg), data...))
}

// Hash returns the hash result from Keccak256(data)
func Hash(data []byte) []byte {
	return crypto.Keccak256(data)
}

// VerifySignature verifies signature using the public key provided.
func VerifySignature(publicKey, message, signature []byte) bool {
	if len(signature) == signatureRSFormatLen+1 {
		// signature in [R || S || V] format is 65 bytes
		//https://bitcoin.stackexchange.com/questions/38351/ecdsa-v-r-s-what-is-v
		signature = signature[0:signatureRSFormatLen]
	}
	// the signature should have the 64 byte [R || S] format
	return secp256k1.VerifySignature(publicKey, Hash(message), signature)

}

// GetSigningKeyPair returns the public and private keys as byte array
func GetSigningKeyPair(pub, priv string) (public, private []byte, err error) {
	privateKey, err := GetPrivateSigningKey(priv)
	if err != nil {
		return nil, nil, errors.New("failed to read private key: %v", err)
	}

	publicKey, err := GetPublicSigningKey(pub)
	if err != nil {
		return nil, nil, errors.New("failed to read public key: %v", err)
	}

	return publicKey, privateKey, nil
}

// GetPrivateSigningKey returns the private key from the file
func GetPrivateSigningKey(fileName string) (key []byte, err error) {
	key, err = utils.ReadKeyFromPemFile(fileName, utils.PrivateKey)
	if err != nil {
		return nil, err
	}
	return key, nil
}

// GetPublicSigningKey returns the public key from the file
func GetPublicSigningKey(fileName string) (key []byte, err error) {
	key, err = utils.ReadKeyFromPemFile(fileName, utils.PublicKey)
	if err != nil {
		return nil, err
	}
	return key, nil
}
