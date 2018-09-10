package utils

import (
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"os"
)

// WriteKeyToPemFile writes encode of key and purpose to the file
func WriteKeyToPemFile(fileName string, keyPurpose string, key []byte) error {
	f, err := os.Create(fileName)
	if err != nil {
		return err
	}

	block := &pem.Block{
		Type:  keyPurpose,
		Bytes: key,
	}
	if err := pem.Encode(f, block); err != nil {
		return err
	}
	f.Close()
	return nil
}

// ReadKeyFromPemFile reads the pem file and returns the key with matching key purpose
func ReadKeyFromPemFile(fileName, keyPurpose string) (key []byte, err error) {
	pemData, err := ioutil.ReadFile(fileName)
	if err != nil {
		return []byte{}, err
	}
	block, _ := pem.Decode(pemData)
	if block == nil {
		return []byte{}, fmt.Errorf("file [%s] is not a valid pem file", fileName)
	}
	if block.Type != keyPurpose {
		return []byte{}, fmt.Errorf("key type mismatch got [%s] but expected [%s]", block.Type, keyPurpose)
	}

	return block.Bytes, nil
}
