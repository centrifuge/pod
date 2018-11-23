package ed25519

import (
	"fmt"

	"github.com/centrifuge/go-centrifuge/utils"
	logging "github.com/ipfs/go-log"
	"github.com/libp2p/go-libp2p-crypto"
	"github.com/libp2p/go-libp2p-peer"
	"golang.org/x/crypto/ed25519"
)

var log = logging.Logger("ed25519")

// GetPublicSigningKey returns the public key from the file
func GetPublicSigningKey(fileName string) (publicKey ed25519.PublicKey, err error) {
	key, err := utils.ReadKeyFromPemFile(fileName, utils.PublicKey)
	if err != nil {
		return nil, fmt.Errorf("failed to read pem file: %v", err)
	}

	return ed25519.PublicKey(key), nil
}

// GetPrivateSigningKey returns the private key from the file
func GetPrivateSigningKey(fileName string) (privateKey ed25519.PrivateKey, err error) {
	key, err := utils.ReadKeyFromPemFile(fileName, utils.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to read pem file: %v", err)
	}

	return ed25519.PrivateKey(key), nil
}

// GetSigningKeyPair returns the public and private key pair
func GetSigningKeyPair(pub, priv string) (publicKey ed25519.PublicKey, privateKey ed25519.PrivateKey, err error) {
	publicKey, err = GetPublicSigningKey(pub)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read public key: %v", err)
	}

	privateKey, err = GetPrivateSigningKey(priv)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read private key: %v", err)
	}

	return publicKey, privateKey, nil
}

// GenerateSigningKeyPair generates ed25519 key pair
func GenerateSigningKeyPair() (publicKey ed25519.PublicKey, privateKey ed25519.PrivateKey) {
	publicKey, privateKey, err := ed25519.GenerateKey(nil)
	if err != nil {
		log.Fatal(err)
	}
	return publicKey, privateKey
}

// PublicKeyToP2PKey returns p2pId from the public key
func PublicKeyToP2PKey(publicKey [32]byte) (p2pID peer.ID, err error) {
	pk, err := crypto.UnmarshalEd25519PublicKey(publicKey[:])
	if err != nil {
		return p2pID, err
	}

	return peer.IDFromPublicKey(pk)
}
