package signingkeys

import (
	"encoding/base64"
	"io/ioutil"

	"golang.org/x/crypto/ed25519"
)

type KeyFiles struct {
	PublicKeyPath  string
	PrivateKeyPath string
}

func GetPublicKey(files KeyFiles) (publicKey ed25519.PublicKey) {
	b64Key, err := ioutil.ReadFile(files.PublicKeyPath)
	if err != nil {
		panic(err)
	}
	var byteKey []byte
	base64.StdEncoding.Decode(byteKey, b64Key)
	publicKey = ed25519.PublicKey(byteKey)
	return
}

func GetPrivateKey(files KeyFiles) (privateKey ed25519.PrivateKey) {
	b64Key, err := ioutil.ReadFile(files.PrivateKeyPath)
	if err != nil {
		panic(err)
	}
	var byteKey []byte
	base64.StdEncoding.Decode(byteKey, b64Key)
	privateKey = ed25519.PrivateKey(byteKey)
	return
}

func GenerateKeypair(publicFileName, privateFileName string) (keyPair KeyFiles) {
	publicKey, privateKey, err := ed25519.GenerateKey(nil)
	if err != nil {
		panic(err)
	}
	keyPair = KeyFiles{publicFileName, privateFileName}

	keyPair.writePrivateKeyToFile(privateKey)
	keyPair.writePublicKeyToFile(publicKey)
	return
}

func (keyPair *KeyFiles) writePublicKeyToFile(publicKey ed25519.PublicKey) {
	b64Key := make([]byte, base64.StdEncoding.EncodedLen(len(publicKey)))

	base64.StdEncoding.Encode(b64Key, publicKey)
	err := ioutil.WriteFile(keyPair.PublicKeyPath, b64Key, 0600)
	if err != nil {
		panic(err)
	}
}

func (keyPair *KeyFiles) writePrivateKeyToFile(privateKey ed25519.PrivateKey) {
	b64Key := make([]byte, base64.StdEncoding.EncodedLen(len(privateKey)))
	base64.StdEncoding.Encode(b64Key, privateKey)
	err := ioutil.WriteFile(keyPair.PrivateKeyPath, b64Key, 0600)
	if err != nil {
		panic(err)
	}
}
