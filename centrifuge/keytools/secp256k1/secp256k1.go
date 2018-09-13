package secp256k1

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"

	"fmt"

	"encoding/base64"

	"github.com/CentrifugeInc/go-centrifuge/centrifuge/config"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/crypto/secp256k1"
	logging "github.com/ipfs/go-log"
)

var log = logging.Logger("signing")

const SignatureRSFormatLen = 64  //64 byte [R || S] format
const SignatureRSVFormatLen = 65 //65 byte [R || S || V] format
const SignatureVPosition = 64
const PrivateKeyLen = 32

func GenerateSigningKeyPair() (publicKey, privateKey []byte) {

	log.Debug("generate secp256k1 keys")
	key, err := ecdsa.GenerateKey(secp256k1.S256(), rand.Reader)
	if err != nil {
		log.Fatal(err)
	}
	publicKey = elliptic.Marshal(secp256k1.S256(), key.X, key.Y)

	privateKey = make([]byte, PrivateKeyLen)
	blob := key.D.Bytes()
	copy(privateKey[PrivateKeyLen-len(blob):], blob)

	return publicKey, privateKey
}

func Sign(message []byte, privateKey []byte) (signature []byte, err error) {
	return secp256k1.Sign(message, privateKey)

}

func SignEthereum(message []byte, privateKey []byte) (signature []byte, err error) {
	// The hash is calculated in Ethereum in the following way
	// keccak256("\x19Ethereum Signed Message:\n"${message length}${message}).
	hash := SignHash(message)
	return Sign(hash, privateKey)
}

func GetAddress(publicKey []byte) string {

	hash := crypto.Keccak256(publicKey[1:])
	address := hash[12:] //address is the last 20 bytes of the hash len(hash) = 20
	return hexutil.Encode(address)
}

func VerifySignatureWithAddress(address, sigHex string, msg []byte) bool {
	fromAddr := common.HexToAddress(address)

	sig, err := hexutil.Decode(sigHex)

	if err != nil {
		log.Error(err.Error())
		return false
	}

	if len(sig) != SignatureRSVFormatLen {
		log.Error("signature must be 65 bytes long")
		return false
	}

	// see implementation in go-ethereum for further details
	// https://github.com/ethereum/go-ethereum/blob/55599ee95d4151a2502465e0afc7c47bd1acba77/internal/ethapi/api.go#L442
	if sig[SignatureVPosition] != 0 && sig[SignatureVPosition] != 1 {
		if sig[SignatureVPosition] != 27 && sig[SignatureVPosition] != 28 {
			log.Error("V value in signature has to be 27 or 28")
			return false
		}
		sig[SignatureVPosition] -= 27 // change V value to 0 or 1
	}

	pubKey, err := crypto.SigToPub(SignHash(msg), sig)
	if err != nil {

		return false
	}

	recoveredAddr := crypto.PubkeyToAddress(*pubKey)

	return fromAddr == recoveredAddr
}

// The hash is calculated as
// keccak256("\x19Ethereum Signed Message:\n"${message length}${message}).
// for further details see
// https://github.com/ethereum/go-ethereum/blob/55599ee95d4151a2502465e0afc7c47bd1acba77/internal/ethapi/api.go#L404

func SignHash(data []byte) []byte {
	msg := fmt.Sprintf("\x19Ethereum Signed Message:\n%d", len(data))
	return crypto.Keccak256(append([]byte(msg), data...))
}

func VerifySignature(publicKey, message, signature []byte) bool {
	if len(signature) == SignatureRSFormatLen+1 {
		// signature in [R || S || V] format is 65 bytes
		//https://bitcoin.stackexchange.com/questions/38351/ecdsa-v-r-s-what-is-v

		signature = signature[0:SignatureRSFormatLen]
	}
	// the signature should have the 64 byte [R || S] format
	return secp256k1.VerifySignature(publicKey, message, signature)

}

// GetIDConfig reads the keys and ID from the config and returns a the Identity config
func GetIDConfig() (identityConfig *config.IdentityConfig, err error) {
	pub, pvk := GetEthAuthKeyFromConfig()
	decodedId, err := base64.StdEncoding.DecodeString(string(config.Config.GetIdentityId()))
	if err != nil {
		return nil, err
	}

	identityConfig = &config.IdentityConfig{
		ID:         decodedId,
		PublicKey:  pub,
		PrivateKey: pvk,
	}
	return
}

// GetEthAuthKeyFromConfig returns the public and private keys as byte array
func GetEthAuthKeyFromConfig() (public, private []byte) {
	pub, priv := config.Config.GetEthAuthKeyPair()
	privateKey, err := GetPrivateEthAuthKey(priv)
	if err != nil {
		log.Error(err)
		return nil, nil
	}
	publicKey, err := GetPublicEthAuthKey(pub)
	if err != nil {
		log.Error(err)
		return nil, nil
	}
	return publicKey, privateKey
}

// GetPrivateEthAuthKey returns the private key from the file
func GetPrivateEthAuthKey(fileName string) (key []byte, err error) {
	key, err = utils.ReadKeyFromPemFile(fileName, utils.PrivateKey)
	if err != nil {
		log.Error(err)
	}
	return
}

// GetPublicEthAuthKey returns the public key from the file
func GetPublicEthAuthKey(fileName string) (key []byte, err error) {
	key, err = utils.ReadKeyFromPemFile(fileName, utils.PublicKey)
	if err != nil {
		log.Error(err)
	}
	return
}
