package tools


import (
	"crypto/rand"
	"math"
	"strings"
	"errors"
)

// Converts a given 32 character long string into a [32]byte
// on error/invalid string lenght returns empty byte array and error
func StringToByte32(input string) (ret [32]byte, err error){
	if len(input) != 32{
		return ret, errors.New("can only work with strings of length 32")
	}
	copy(ret[:],input)
	return
}

// Converts a given [32]byte into a 32 char length string
// on error/invalid input, returns empty string and error
func Byte32ToString(input [32]byte) (ret string, err error){
	ret = string(input[:32])
	return
}


func RandomString32() (ret string){
	b := RandomByte32()
	ret = string(b[:32])
	return
}


func RandomByte32() (out [32]byte) {
	r := make([]byte, 32)
	_, err := rand.Read(r)
	// Note that err == nil only if we read len(b) bytes.
	if err != nil {
		panic(err)
	}
	copy(out[:], r[:32])
	return
}

func StrPadHex32(input string) string{
	return StrPad(input, 32, "0","LEFT")
}


// StrPad returns the input string padded on the left, right or both sides using padType to the specified padding length padLength.
//
// Example:
// input := "Codes";
// StrPad(input, 10, " ", "RIGHT")        // produces "Codes     "
// StrPad(input, 10, "-=", "LEFT")        // produces "=-=-=Codes"
// StrPad(input, 10, "_", "BOTH")         // produces "__Codes___"
// StrPad(input, 6, "___", "RIGHT")       // produces "Codes_"
// StrPad(input, 3, "*", "RIGHT")         // produces "Codes"
func StrPad(input string, padLength int, padString string, padType string) string {
	var output string

	inputLength := len(input)
	padStringLength := len(padString)

	if inputLength >= padLength {
		return input
	}

	repeat := math.Ceil(float64(1) + (float64(padLength-padStringLength))/float64(padStringLength))

	switch padType {
	case "RIGHT":
		output = input + strings.Repeat(padString, int(repeat))
		output = output[:padLength]
	case "LEFT":
		output = strings.Repeat(padString, int(repeat)) + input
		output = output[len(output)-padLength:]
	case "BOTH":
		length := (float64(padLength - inputLength)) / float64(2)
		repeat = math.Ceil(length / float64(padStringLength))
		output = strings.Repeat(padString, int(repeat))[:int(math.Floor(float64(length)))] + input + strings.Repeat(padString, int(repeat))[:int(math.Ceil(float64(length)))]
	}

	return output
}