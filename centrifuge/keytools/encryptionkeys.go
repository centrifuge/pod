package keytools

import (
	"io/ioutil"
	"github.com/spf13/viper"
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/elliptic"
	"github.com/ethereum/go-ethereum/crypto/secp256k1"
	"github.com/ethereum/go-ethereum/common/math"

)


func GetPublicEncryptionKey(fileName string) (publicKey []byte) {
	key, err := ioutil.ReadFile(fileName)
	if err != nil {
		panic(err)
	}
	return key
}

func GetPrivateEncryptionKey(fileName string) (privateKey []byte) {
	key, err := ioutil.ReadFile(fileName)
	if err != nil {
		panic(err)
	}
	return key
}

func GetEncryptionKeysFromConfig() (publicKey, privateKey []byte) {
	publicKey = GetPublicEncryptionKey(viper.GetString("keys.encryption.publicKey"))
	privateKey = GetPrivateEncryptionKey(viper.GetString("keys.encryption.privateKey"))
	return
}

func GenerateEncryptionKeypair(publicFileName, privateFileName string) (publicKey, privateKey []byte) {
	key, err := ecdsa.GenerateKey(secp256k1.S256(), rand.Reader)
	if err != nil {
		panic(err)
	}
	publicKey = elliptic.Marshal(secp256k1.S256(), key.X, key.Y)
	privateKey = math.PaddedBigBytes(key.D, 32)

	writeKeyToFile(privateFileName, privateKey)
	writeKeyToFile(publicFileName, publicKey)
	return
}


