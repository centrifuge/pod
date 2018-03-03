package keytools

import (
	"io/ioutil"

	"github.com/spf13/viper"
	"golang.org/x/crypto/ed25519"

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

