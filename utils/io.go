package utils

import (
	"encoding/pem"
	"io/ioutil"
	"os"

	"github.com/centrifuge/go-centrifuge/errors"
)

// WriteKeyToPemFile writes encode of key and purpose to the file
func WriteKeyToPemFile(fileName string, keyPurpose string, key []byte) error {
	// we are going with 0622 so that when umask is applied, we will get 0600 file permission
	f, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0622)
	if err != nil {
		return err
	}

	defer f.Close()

	block := &pem.Block{
		Type:  keyPurpose,
		Bytes: key,
	}
	if err := pem.Encode(f, block); err != nil {
		return err
	}

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
		return []byte{}, errors.New("file [%s] is not a valid pem file", fileName)
	}
	if block.Type != keyPurpose {
		return []byte{}, errors.New("key type mismatch got [%s] but expected [%s]", block.Type, keyPurpose)
	}

	return block.Bytes, nil
}
