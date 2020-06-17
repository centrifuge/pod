package byteutils

import (
	"bytes"
	"math/big"
	"sort"
	"strings"

	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

// AddZeroBytesSuffix appends zero bytes such that result byte length == required
func AddZeroBytesSuffix(data []byte, required int) []byte {
	if len(data) >= required {
		return data
	}

	tba := required - len(data)
	return append(data, make([]byte, tba)...)
}

// RemoveZeroBytesSuffix removes zero bytes appended to the end.
func RemoveZeroBytesSuffix(data []byte) []byte {
	if len(data) < 1 {
		return data
	}

	for i := len(data) - 1; i >= 0; i-- {
		if data[i] != 0 {
			return data[:i+1]
		}
	}

	return nil
}

// IntBytesFromString return the integer base 10 string in bytes.
func IntBytesFromString(s string) ([]byte, error) {
	s = strings.TrimSpace(s)
	if len(s) < 1 {
		return nil, nil
	}

	d, ok := new(big.Int).SetString(s, 10)
	if !ok {
		return nil, errors.New("invalid integer string")
	}

	return d.Bytes(), nil
}

// ContainsBytesInSlice returns bool if byte slice is contained in input
func ContainsBytesInSlice(slice [][]byte, b []byte) bool {
	for _, s := range slice {
		if utils.IsSameByteSlice(s, b) {
			return true
		}
	}

	return false
}

// RemoveBytesFromSlice removes bytes b from the slice of bytes
// Note: all duplicates of b in slice will be removed.
// Note: if thee bytes b doesn't exist, same slice is returned
func RemoveBytesFromSlice(slice [][]byte, b []byte) [][]byte {
	var res [][]byte
	for _, s := range slice {
		s := s
		if bytes.Equal(s, b) {
			continue
		}

		res = append(res, s)
	}

	return res
}

// SetBit sets the bit at pos in the given byte.
func SetBit(n byte, pos uint) byte {
	n |= 1 << pos
	return n
}

// ClearBit clears the bit at pos in n.
func ClearBit(n byte, pos uint) byte {
	mask := ^(1 << pos)
	n &= byte(mask)
	return n
}

// IsBitSet checks if the bit at position is set
func IsBitSet(n byte, pos uint) bool {
	val := n & (1 << pos)
	return val > 0
}

// BytesArray is a alias for 32 byte alice
type BytesArray [][32]byte

// Len returns the length of the slice
func (a BytesArray) Len() int {
	return len(a)
}

// Less returns true if i'th item is less than j'th else false
func (a BytesArray) Less(i, j int) bool {
	switch bytes.Compare(a[i][:], a[j][:]) {
	case -1:
		return true
	default:
		return false
	}
}

// Swap swaps the i, j values with in the array.
func (a BytesArray) Swap(i, j int) {
	a[j], a[i] = a[i], a[j]
}

// SortByte32Slice sorts the byte32 slices in ascending order.
func SortByte32Slice(arr [][32]byte) [][32]byte {
	ba := BytesArray(arr)
	sort.Sort(ba)
	return ba
}

// HexBytes is of type bytes with JSON hex string marshaller.
type HexBytes []byte

// MarshalJSON marshall bytes to hex.
func (h HexBytes) MarshalJSON() ([]byte, error) {
	str := "0x"
	if len(h) > 0 {
		str = hexutil.Encode(h)
	}

	str = "\"" + str + "\""
	return []byte(str), nil
}

// UnmarshalJSON unmarshals hex string to bytes.
func (h *HexBytes) UnmarshalJSON(data []byte) error {
	str := string(data)
	str = strings.TrimSpace(strings.Trim(str, "\""))
	d, err := hexutil.Decode(str)
	if err != nil {
		return err
	}

	*h = HexBytes(d)
	return nil
}

// ToHexByteSlice converts slicee of slice of bytes to slice of HexBytes
func ToHexByteSlice(l [][]byte) (hb []HexBytes) {
	for _, b := range l {
		b := b
		hb = append(hb, b)
	}

	return hb
}

// String returns HexBytes in string format.
func (h HexBytes) String() string {
	return hexutil.Encode(h)
}

// Bytes returns a copy of the HexBytes.
func (h HexBytes) Bytes() []byte {
	if len(h) < 1 {
		return nil
	}

	d := make([]byte, len(h))
	copy(d, h[:])
	return d
}

// OptionalHex can have empty hex string and doesn't error.
type OptionalHex struct {
	HexBytes
}

// UnmarshalJSON unmarshalls hex encoded string to bytes.
// if the data is an empty string, it decoded value is empty bytes.
func (o *OptionalHex) UnmarshalJSON(data []byte) error {
	str := string(data)
	str = strings.TrimSpace(strings.Trim(str, "\""))
	if str == "" {
		o.HexBytes = nil
		return nil
	}

	d, err := hexutil.Decode(str)
	if err != nil {
		return err
	}

	*o = OptionalHex{HexBytes(d)}
	return nil
}

// CutFromSlice returns a new slice without the value at idx
func CutFromSlice(slice [][]byte, idx int) [][]byte {
	if idx >= len(slice) {
		return slice
	}

	var ns [][]byte
	for i, v := range slice {
		if i == idx {
			continue
		}

		ns = append(ns, v)
	}

	return ns
}
