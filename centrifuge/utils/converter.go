package utils

import (
	"encoding/hex"
)

const HexPrefix = "0x"

func HexToByteArray(hexString string) []byte {

	if hexString[0:len(HexPrefix)] == HexPrefix {
		hexString = hexString[len(HexPrefix):]
	}
	byteArray, err := hex.DecodeString(hexString)

	if err != nil {
		log.Fatal(err)
	}

	return byteArray

}

func ByteArrayToHex(byteArray []byte) string {
	return HexPrefix + hex.EncodeToString(byteArray)

}
