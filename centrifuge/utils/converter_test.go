package utils

import (
	"testing"

	"github.com/magiconair/properties/assert"
)

func TestHexToByteArray(t *testing.T) {
	testHexWithPrefix := "0xd77c534aed04d7ce34cd425073a033db4fbe6a9d" //with prefix
	byteArrayFromPrefixInput := HexToByteArray(testHexWithPrefix)

	testHex := "d77c534aed04d7ce34cd425073a033db4fbe6a9d" //without prefix
	byteArray := HexToByteArray(testHex)

	assert.Equal(t, byteArrayFromPrefixInput, byteArray, "converting hex string into byte array has problems with 0x prefix")

}

func TestByteArrayToHex(t *testing.T) {
	testHex := "0xd77c534aed04d7ce34cd425073a033db4fbe6a9d" //with prefix
	byteArray := HexToByteArray(testHex)

	assert.Equal(t, testHex, ByteArrayToHex(byteArray), "converting byte array to hex didn't work correctly")

}
