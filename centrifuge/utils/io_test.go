// +build unit

package utils

import (
	"os"
	"testing"

	"github.com/centrifuge/go-centrifuge/centrifuge/tools"
	"github.com/stretchr/testify/assert"
)

func TestWriteAndReadPemFile(t *testing.T) {

	randomBytes := tools.RandomSlice(32)

	testFileName := "test.file"
	err := WriteKeyToPemFile(testFileName, PrivateKey, randomBytes)
	assert.Nil(t, err, "writing file failed")

	fileContent, err := ReadKeyFromPemFile(testFileName, PrivateKey)
	assert.Nil(t, err, "failed to read written file correctly")
	assert.Equal(t, randomBytes, fileContent, "content of the written file is not correct")

	fileContent, err = ReadKeyFromPemFile(testFileName, PublicKey)
	assert.Error(t, err, "should produce an error because key file contains a private key ")

	os.Remove(testFileName)

}
