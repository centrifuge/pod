package byteutils

import (
	"errors"
	"math/big"
	"strings"
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
