package signingkeys

import (
	"io/ioutil"

	"github.com/spf13/viper"
	"golang.org/x/crypto/ed25519"
)

func GetPublicKey(fileName string) (publicKey ed25519.PublicKey) {
	key, err := ioutil.ReadFile(fileName)
	if err != nil {
		panic(err)
	}
	publicKey = ed25519.PublicKey(key)
	return
}

func GetPrivateKey(fileName string) (privateKey ed25519.PrivateKey) {
	key, err := ioutil.ReadFile(fileName)
	if err != nil {
		panic(err)
	}
	privateKey = ed25519.PrivateKey(key)
	return
}

func GetKeysFromConfig() (publicKey ed25519.PublicKey, privateKey ed25519.PrivateKey) {
	publicKey = GetPublicKey(viper.GetString("witness.publicKey"))
	privateKey = GetPrivateKey(viper.GetString("witness.privateKey"))
	return
}

func GenerateKeypair(publicFileName, privateFileName string) (publicKey ed25519.PublicKey, privateKey ed25519.PrivateKey) {
	publicKey, privateKey, err := ed25519.GenerateKey(nil)
	if err != nil {
		panic(err)
	}
	writeKeyToFile(privateFileName, privateKey)
	writeKeyToFile(publicFileName, publicKey)
	return
}

func writeKeyToFile(fileName string, key []byte) {
	err := ioutil.WriteFile(fileName, key, 0600)
	if err != nil {
		panic(err)
	}
}
