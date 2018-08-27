package utils

import (
	"encoding/hex"
)

const HEX_PREFIX = "0x"

func HexToByteArray(hexString string) ([]byte){

	if(hexString[0:len(HEX_PREFIX)] == HEX_PREFIX){
		hexString = hexString[len(HEX_PREFIX):]
	}
	byteArray, err := hex.DecodeString(hexString)

	if(err != nil){
		log.Fatal(err)
	}

	return byteArray

}

func ByteArrayToHex(byteArray []byte) (string){
	return HEX_PREFIX+hex.EncodeToString(byteArray)

}
