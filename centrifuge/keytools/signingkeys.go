package keytools

import (
	"io/ioutil"

	"github.com/spf13/viper"
	"golang.org/x/crypto/ed25519"
	mh "github.com/multiformats/go-multihash"
	"github.com/libp2p/go-libp2p-peer"
	"github.com/libp2p/go-libp2p-crypto"
)

func GetPublicSigningKey(fileName string) (publicKey ed25519.PublicKey) {
	key, err := ioutil.ReadFile(fileName)
	if err != nil {
		panic(err)
	}
	publicKey = ed25519.PublicKey(key)
	return
}

func GetPrivateSigningKey(fileName string) (privateKey ed25519.PrivateKey) {
	key, err := ioutil.ReadFile(fileName)
	if err != nil {
		panic(err)
	}
	privateKey = ed25519.PrivateKey(key)
	return
}

func GetSigningKeysFromConfig() (publicKey ed25519.PublicKey, privateKey ed25519.PrivateKey) {
	publicKey = GetPublicSigningKey(viper.GetString("keys.signing.publicKey"))
	privateKey = GetPrivateSigningKey(viper.GetString("keys.signing.privateKey"))
	return
}

func GenerateSigningKeypair(publicFileName, privateFileName string) (publicKey ed25519.PublicKey, privateKey ed25519.PrivateKey) {
	publicKey, privateKey, err := ed25519.GenerateKey(nil)
	if err != nil {
		panic(err)
	}
	writeKeyToFile(privateFileName, privateKey)
	writeKeyToFile(publicFileName, publicKey)
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

