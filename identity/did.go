package identity

import (
	"bytes"
	"encoding/gob"
	"strings"

	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

const (

	// ErrMalformedAddress standard error for malformed address
	ErrMalformedAddress = errors.Error("malformed address provided")

	// ErrInvalidDIDLength must be used with invalid bytelength when attempting to convert to a DID
	ErrInvalidDIDLength = errors.Error("invalid DID length")
)

func init() {
	gob.Register(DID{})
}

// DID stores the identity address of the user
type DID [DIDLength]byte

// DIDLength contains the length of a DID
const DIDLength = 32

// MarshalJSON marshals DID to json bytes.
func (d DID) MarshalJSON() ([]byte, error) {
	str := "\"" + d.ToHexString() + "\""
	return []byte(str), nil
}

// UnmarshalJSON loads json bytes to DID
func (d *DID) UnmarshalJSON(data []byte) error {
	dx, err := NewDIDFromString(strings.Trim(string(data), "\""))
	if err != nil {
		return err
	}
	*d = dx
	return nil
}

// ToHexString returns the DID as a hex string
func (d DID) ToHexString() string {
	return hexutil.Encode(d[:])
}

// Equal checks if d == other
func (d DID) Equal(other DID) bool {
	return bytes.Equal(d[:], other[:])
}

// NewDIDFromString returns a DID based on a hex string
func NewDIDFromString(address string) (DID, error) {
	b, err := hexutil.Decode(address)

	if err != nil {
		return DID{}, err
	}

	return NewDIDFromBytes(b)
}

// NewDIDFromBytes returns a DID based on a bytes input
func NewDIDFromBytes(bAddr []byte) (DID, error) {
	if len(bAddr) != DIDLength {
		return DID{}, ErrInvalidDIDLength
	}

	var b [32]byte

	copy(b[:], bAddr)

	return b, nil
}

// ValidateDIDBytes validates a centrifuge ID given as bytes
func ValidateDIDBytes(givenDID []byte, did DID) error {
	calcdid, err := NewDIDFromBytes(givenDID)
	if err != nil {
		return err
	}
	if !did.Equal(calcdid) {
		return errors.New("provided bytes doesn't match centID")
	}

	return nil
}
