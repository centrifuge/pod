package utils

import (
	"crypto/rand"
	"errors"
	"math/big"

	"github.com/centrifuge/gocelery"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

// ContainsBigIntInSlice checks if value is present in list.
func ContainsBigIntInSlice(value *big.Int, list []*big.Int) bool {
	for _, v := range list {
		if v.Cmp(value) == 0 {
			return true
		}
	}
	return false
}

// SliceToByte32 converts a 32 byte slice to an array. Will thorw error if the slice is too long
func SliceToByte32(in []byte) (out [32]byte, err error) {
	if len(in) > 32 {
		return [32]byte{}, errors.New("input exceeds length of 32")
	}
	copy(out[:], in)
	return out, nil
}

// IsEmptyAddress checks if the addr is empty.
func IsEmptyAddress(addr common.Address) bool {
	return addr.Hex() == "0x0000000000000000000000000000000000000000"
}

// SliceOfByteSlicesToHexStringSlice converts the given slice of byte slices to a hex encoded string array with 0x prefix
func SliceOfByteSlicesToHexStringSlice(byteSlices [][]byte) []string {
	hexArr := make([]string, len(byteSlices))
	for i, b := range byteSlices {
		hexArr[i] = hexutil.Encode(b)
	}
	return hexArr
}

// Byte32ToSlice converts a [32]bytes to an unbounded byte array
func Byte32ToSlice(in [32]byte) []byte {
	if IsEmptyByte32(in) {
		return []byte{}
	}

	return in[:]
}

// CheckMultiple32BytesFilled takes multiple []byte slices and ensures they are all of length 32 and don't contain all 0x0 bytes.
func CheckMultiple32BytesFilled(b []byte, bs ...[]byte) bool {
	bs = append(bs, b)
	for _, v := range bs {
		if IsEmptyByteSlice(v) || len(v) != 32 {
			return false
		}
	}
	return true
}

// RandomSlice returns a randomly filled byte array with length of given size
func RandomSlice(size int) (out []byte) {
	r := make([]byte, size)
	_, err := rand.Read(r)
	// Note that err == nil only if we read len(b) bytes.
	if err != nil {
		panic(err)
	}
	return r
}

// RandomByte32 returns a randomly filled byte array with length of 32
func RandomByte32() (out [32]byte) {
	r := RandomSlice(32)
	copy(out[:], r[:32])
	return
}

// IsEmptyByte32 checks if the source is empty.
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

// IsSameByteSlice checks if a and b contains same bytes.
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

// ByteSliceToBigInt convert bute slices to big.Int (bigendian)
func ByteSliceToBigInt(slice []byte) *big.Int {
	bi := new(big.Int)
	bi.SetBytes(slice)
	return bi
}

// ByteFixedToBigInt convert arbitrary length byte arrays to big.Int
func ByteFixedToBigInt(bytes []byte, size int) *big.Int {
	bi := new(big.Int)
	bi.SetBytes(bytes[:size])
	return bi
}

// SimulateJSONDecodeForGocelery encodes and decodes the kwargs.
func SimulateJSONDecodeForGocelery(kwargs map[string]interface{}) (map[string]interface{}, error) {
	t1 := gocelery.TaskMessage{Kwargs: kwargs}
	encoded, err := t1.Encode()
	if err != nil {
		return nil, err
	}

	t2, err := gocelery.DecodeTaskMessage(encoded)
	return t2.Kwargs, err
}

// IsValidByteSliceForLength checks if the len(slice) == length.
func IsValidByteSliceForLength(slice []byte, length int) bool {
	return len(slice) == length
}
