package signingkeys

import (
	"encoding/base64"
	"io/ioutil"

	"golang.org/x/crypto/ed25519"
	"github.com/spf13/viper"
)


func GetPublicKey(fileName string) (publicKey ed25519.PublicKey) {
	b64Key, err := ioutil.ReadFile(fileName)
	if err != nil {
		panic(err)
	}
	var byteKey []byte
	base64.StdEncoding.Decode(byteKey, b64Key)
	publicKey = ed25519.PublicKey(byteKey)
	return
}

func GetPrivateKey(fileName string) (privateKey ed25519.PrivateKey) {
	b64Key, err := ioutil.ReadFile(fileName)
	if err != nil {
		panic(err)
	}
	var byteKey []byte
	base64.StdEncoding.Decode(byteKey, b64Key)
	privateKey = ed25519.PrivateKey(byteKey)
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
	b64Key := make([]byte, base64.StdEncoding.EncodedLen(len(key)))

	base64.StdEncoding.Encode(b64Key, key)
	err := ioutil.WriteFile(fileName, b64Key, 0600)
	if err != nil {
		panic(err)
	}
}
