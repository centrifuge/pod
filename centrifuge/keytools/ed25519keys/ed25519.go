package ed25519keys

import (
	"github.com/centrifuge/go-centrifuge/centrifuge/config"
	"github.com/centrifuge/go-centrifuge/centrifuge/utils"
	logging "github.com/ipfs/go-log"
	"github.com/libp2p/go-libp2p-crypto"
	"github.com/libp2p/go-libp2p-peer"
	"golang.org/x/crypto/ed25519"
)

var log = logging.Logger("ed25519")

// GetPublicSigningKey returns the public key from the file
func GetPublicSigningKey(fileName string) (publicKey ed25519.PublicKey) {
	key, err := utils.ReadKeyFromPemFile(fileName, utils.PublicKey)

	if err != nil {
		log.Fatal(err)
	}
	publicKey = ed25519.PublicKey(key)
	return
}

// GetPrivateSigningKey returns the private key from the file
func GetPrivateSigningKey(fileName string) (privateKey ed25519.PrivateKey) {
	key, err := utils.ReadKeyFromPemFile(fileName, utils.PrivateKey)
	if err != nil {
		log.Fatal(err)
	}
	privateKey = ed25519.PrivateKey(key)
	return
}

// GetSigningKeyPairFromConfig returns the public and private key pair from the config
func GetSigningKeyPairFromConfig() (publicKey ed25519.PublicKey, privateKey ed25519.PrivateKey) {
	pub, priv := config.Config.GetSigningKeyPair()
	publicKey = GetPublicSigningKey(pub)
	privateKey = GetPrivateSigningKey(priv)
	return
}

// GenerateSigningKeyPair generates ed25519 key pair
func GenerateSigningKeyPair() (publicKey ed25519.PublicKey, privateKey ed25519.PrivateKey) {
	publicKey, privateKey, err := ed25519.GenerateKey(nil)
	if err != nil {
		log.Fatal(err)
	}
	return
}

// PublicKeyToP2PKey returns p2pId from the public key
func PublicKeyToP2PKey(publicKey [32]byte) (p2pId peer.ID, err error) {
	pk, err := crypto.UnmarshalEd25519PublicKey(publicKey[:])
	if err != nil {
		return "", err
	}

	p2pId, err = peer.IDFromPublicKey(pk)
	if err != nil {
		return "", err
	}
	return
}

// GetIDConfig reads the keys and ID from the config and returns a the Identity config
func GetIDConfig() (identityConfig *config.IdentityConfig, err error) {
	pk, pvk := GetSigningKeyPairFromConfig()
	centID, err := config.Config.GetIdentityID()
	if err != nil {
		return nil, err
	}

	identityConfig = &config.IdentityConfig{
		ID:         centID,
		PublicKey:  pk,
		PrivateKey: pvk,
	}
	return
}
