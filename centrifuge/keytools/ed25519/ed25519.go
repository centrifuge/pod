package ed25519

import (
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/config"
	"github.com/libp2p/go-libp2p-crypto"
	"github.com/libp2p/go-libp2p-peer"
	mh "github.com/multiformats/go-multihash"
	"golang.org/x/crypto/ed25519"
	logging "github.com/ipfs/go-log"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/keytools/io"
)

var log = logging.Logger("ed25519")

const (
	PUBLIC_KEY  = "PUBLIC KEY"
	PRIVATE_KEY = "PRIVATE KEY"
)

func GetPublicSigningKey(fileName string) (publicKey ed25519.PublicKey) {
	key, err := io.ReadKeyFromPemFile(fileName, PUBLIC_KEY)

	if err != nil {
		log.Fatal(err)
	}
	publicKey = ed25519.PublicKey(key)
	return
}

func GetPrivateSigningKey(fileName string) (privateKey ed25519.PrivateKey) {
	key, err := io.ReadKeyFromPemFile(fileName, PRIVATE_KEY)
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

func GenerateSigningKeyPair() (publicKey ed25519.PublicKey, privateKey ed25519.PrivateKey) {

	log.Debug("sign ED25519")
	publicKey, privateKey, err := ed25519.GenerateKey(nil)

	if err != nil {
		log.Fatal(err)
	}
	return
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
