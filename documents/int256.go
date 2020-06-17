package documents

import (
	"math/big"
	"regexp"
	"strings"

	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/utils/byteutils"
	"github.com/ethereum/go-ethereum/common/math"
)

// signByteIdx is the index of byte that contains the sign bit of the int256 when represented as 32 bytes
const signByteIdx = 0

// validInt regex for only allowing integers
var validInt = regexp.MustCompile(`^[-+]?\d+$`)

// Int256 represents a signed 256 bit integer
type Int256 struct {
	v big.Int
}

// MarshalJSON marshals decimal to json bytes.
func (i *Int256) MarshalJSON() ([]byte, error) {
	return i.v.MarshalJSON()
}

// UnmarshalJSON loads json bytes to decimal
func (i *Int256) UnmarshalJSON(data []byte) error {
	v := new(big.Int)
	err := v.UnmarshalJSON(data)
	if err != nil {
		return err
	}

	i.v = *v
	return nil
}

// String converts Int256 to string
func (i *Int256) String() string {
	return i.v.String()
}

// NewInt256 creates a new Int256 given a string
func NewInt256(n string) (*Int256, error) {
	n = strings.TrimSpace(n)
	if !validInt.MatchString(n) {
		return nil, errors.NewTypedError(ErrInvalidInt256, errors.New("probably a decimal value: %s", n))
	}

	nn := new(big.Int)
	nn, ok := nn.SetString(n, 10)
	if !ok {
		return nil, errors.NewTypedError(ErrInvalidInt256, errors.New("probably an arbitrary string: %s", n))
	}

	if !isValidInt256(*nn) {
		return nil, errors.NewTypedError(ErrInvalidInt256, errors.New("value: %s", n))
	}
	return &Int256{*nn}, nil
}

// Int256FromBytes converts the a big endian byte slice to an Int256
func Int256FromBytes(b []byte) (*Int256, error) {
	if len(b) != 32 {
		return nil, errors.NewTypedError(ErrInvalidInt256, errors.New("value: %x", b))
	}

	// check and record the sign bit
	neg := false
	if byteutils.IsBitSet(b[signByteIdx], 7) {
		neg = true
	}
	// ignore the sign bit
	b[signByteIdx] = byteutils.ClearBit(b[signByteIdx], 7)

	nn := new(big.Int)
	nn.SetBytes(b[:])
	if !isValidInt256(*nn) {
		return nil, errors.NewTypedError(ErrInvalidInt256, errors.New("value: %s", nn.String()))
	}

	// negative
	if neg {
		nn = nn.Neg(nn)
	}

	return &Int256{*nn}, nil
}

// Bytes returns the big endian 32 byte representation of this int256. First bit of LSB(least significant byte) is used as the sign bit.
func (i *Int256) Bytes() [32]byte {
	var b [32]byte
	// no of bits in i.v.Bytes() <= 255
	// if its less, pad the number in big endian order and copy to the 32 byte array
	copy(b[:], math.PaddedBigBytes(&i.v, 32))

	//  set the sign bit
	b[signByteIdx] = byteutils.ClearBit(b[signByteIdx], 7)
	if i.v.Sign() == -1 {
		b[signByteIdx] = byteutils.SetBit(b[signByteIdx], 7)
	}
	return b
}

// Equals checks if the given two Int256s are equal
func (i *Int256) Equals(o *Int256) bool {
	return i.v.Cmp(&o.v) == 0
}

func isValidInt256(n big.Int) bool {
	x := n.BitLen()
	if x > 256 {
		return false
	}

	// check max
	two := big.NewInt(2)
	exp := two.Exp(two, big.NewInt(255), nil)
	maxI256 := exp.Sub(exp, big.NewInt(1))
	return n.Abs(&n).Cmp(maxI256) <= 0
}

// Add sets i to the sum x+y and returns i
func (i *Int256) Add(x *Int256, y *Int256) (*Int256, error) {
	i.v.Add(&x.v, &y.v)
	if !isValidInt256(i.v) {
		return nil, errors.NewTypedError(ErrInvalidInt256, errors.New("value: %s", &i.v))
	}
	return i, nil
}

// Cmp compares i and y and returns:
//
//   -1 if i <  y
//    0 if i == y
//   +1 if i >  y
//
func (i *Int256) Cmp(y *Int256) int {
	return i.v.Cmp(&y.v)
}

// Inc increments i by one
func (i *Int256) Inc() (*Int256, error) {
	i.v.Add(&i.v, big.NewInt(1))
	if !isValidInt256(i.v) {
		return nil, errors.NewTypedError(ErrInvalidInt256, errors.New("value: %s", &i.v))
	}
	return i, nil
}
