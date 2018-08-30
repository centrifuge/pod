package tools

import (
	"crypto/rand"
	"errors"
	"math/big"
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
func CheckMultiple32BytesFilled(bs ...[]byte) bool {
	for _, v := range bs {
		if IsEmptyByteSlice(v) || len(v) != 32 {
			return false
		}
	}
	return true
}

// RandomSliceN returns a randomly filled byte array with length of n
func RandomSliceN(n int) (out []byte) {
	r := make([]byte, n)
	_, err := rand.Read(r)
	// Note that err == nil only if we read len(b) bytes.
	if err != nil {
		panic(err)
	}
	return r
}

// RandomByte32 returns a randomly filled byte array with length of 32
func RandomByte32() (out [32]byte) {
	r := RandomSliceN(32)
	copy(out[:], r[:32])
	return
}

func IsEmptyByte32(source [32]byte) bool {
	sl := make([]byte, 32)
	copy(sl, source[:32])
	return IsEmptyByteSlice(sl)
}

// IsEmptyByteSlice checks if the provided slice is empty
// returns true if:
// s == nil
// every element is == 0
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

func ByteSliceToBigInt(slice []byte) *big.Int {
	bi := new(big.Int)
	bi.SetBytes(slice)
	return bi
}

func Byte48ToBigInt(bytesArr [48]byte) *big.Int {
	bi := new(big.Int)
	bi.SetBytes(bytesArr[:])
	return bi
}

