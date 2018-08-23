package keytools

import (
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/config"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/keytools/secp256k1"
	logging "github.com/ipfs/go-log"
	"github.com/libp2p/go-libp2p-crypto"
	"github.com/libp2p/go-libp2p-peer"
	mh "github.com/multiformats/go-multihash"
	"golang.org/x/crypto/ed25519"
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

func GetPublicSigningKey(fileName string) (publicKey ed25519.PublicKey) {
	key, err := readKeyFromPemFile(fileName, PUBLIC_KEY)
	if err != nil {
		log.Fatal(err)
	}
	publicKey = ed25519.PublicKey(key)
	return
}

func GetPrivateSigningKey(fileName string) (privateKey ed25519.PrivateKey) {
	key, err := readKeyFromPemFile(fileName, PRIVATE_KEY)
	if err != nil {
		log.Fatal(err)
	}
	privateKey = ed25519.PrivateKey(key)
	return
}

func GetSigningKeyPairFromConfig() (publicKey ed25519.PublicKey, privateKey ed25519.PrivateKey) {
	pub, priv := config.Config.GetSigningKeyPair()
	publicKey = GetPublicSigningKey(pub)
	privateKey = GetPrivateSigningKey(priv)
	return
}

func GenerateSigningKeyPairED25519 () (publicKey ed25519.PublicKey, privateKey ed25519.PrivateKey) {

	log.Debug("sign ED25519")
	publicKey, privateKey, err := ed25519.GenerateKey(nil)

	if err != nil {
		log.Fatal(err)
	}
	return
}

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

func PublicKeyToP2PKey(publicKey [32]byte) (p2pId peer.ID, err error) {
	// Taken from peer.go#IDFromPublicKey#L189
	// TODO As soon as this is merged: https://github.com/libp2p/go-libp2p-kad-dht/pull/129 we can get rid of this function
	// and only do:
	// pk, err := crypto.UnmarshalEd25519PublicKey(publicKey[:])
	// pid, error := IDFromPublicKey(pk)
	pk, err := crypto.UnmarshalEd25519PublicKey(publicKey[:])
	bpk, err := pk.Bytes()
	hash, err := mh.Sum(bpk[:], mh.SHA2_256, -1)
	if err != nil {
		return "", err
	}
	//
	p2pId = peer.ID(hash)
	return
}
