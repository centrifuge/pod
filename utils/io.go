package utils

import (
	"encoding/pem"
	"io/ioutil"
	"os"

	"github.com/centrifuge/pod/errors"
)

// WriteKeyToPemFile writes encode of key and purpose to the file
func WriteKeyToPemFile(fileName string, keyPurpose string, key []byte) (err error) {
	f, err := os.Create(fileName)
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			return
		}

		err = f.Close()
	}()

	block := &pem.Block{
		Type:  keyPurpose,
		Bytes: key,
	}
	if err := pem.Encode(f, block); err != nil {
		return err
	}

	return f.Chmod(os.FileMode(0600))
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
