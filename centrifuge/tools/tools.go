package tools

import (
	"crypto/rand"
	"errors"
	"fmt"
	"math"
	"strings"
)

// SliceToByte32 converts a 32 byte slice to an array. Will thorw error if the slice is too long
func SliceToByte32(in []byte) (out [32]byte, err error) {
	if len(in) > 32 {
		return [32]byte{}, errors.New("input exceeds length of 32")
	}
	copy(out[:], in)
	return
}

// Byte32ToSlice converts a [32]bytes to an unbounded byte array
func Byte32ToSlice(in [32]byte) []byte {
	if IsEmptyByte32(in) {
		return []byte{}
	} else {
		return in[:]
	}
}

// Check32BytesFilled takes multiple []byte slices and ensures they are all of length 32 and don't contain all 0x0 bytes.
func CheckMultiple32BytesFilled(b []byte, bs ...[]byte) bool {
	if IsEmptyByteSlice(b) || len(b) != 32 {
		return false
	}
	for _, v := range bs {
		if !IsEmptyByteSlice(v) || len(v) != 32 {
			return false
		}
	}
	return true
}

func RandomString32() (ret string) {
	b := RandomByte32()
	ret = string(b[:32])
	return
}

// RandomByte32 returns a randomly filled byte array with length of 32
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

func IsEmptyByte32(source [32]byte) bool {
	sl := make([]byte, 32)
	copy(sl, source[:32])
	return IsEmptyByteSlice(sl)
}

func IsEmptyByteSlice(s []byte) bool {
	if s == nil {
		return true
	}
	for _, v := range s {
		if v != 0 {
			return false
		}
	}
	return true
}

func IsSameByteSlice(a []byte, b []byte) bool {
	if a == nil && b == nil {
		return true
	}

	if a == nil || b == nil {
		return false
	}

	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}

func StrPadHex32(input string) string {
	return StrPad(input, 32, "0", "LEFT")
}

// CheckLen32 is used to validate that the given val is 32 characters long. If not, it returns an error with the error
// message of `errorMessage`
func CheckLen32(val string, errorMessage string) error {
	if len(val) != 32 {
		return errors.New(fmt.Sprintf(errorMessage, val))
	}
	return nil
}

// CheckBytesLen32 is used to validate that a given byte slice is of length 32, if not it returns an error with the error
// message of `errorMessage`
func CheckBytesLen32(val []byte, errorMessage string) error {
	if len(val) != 32 {
		return errors.New(fmt.Sprintf(errorMessage, val))
	}
	return nil
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
