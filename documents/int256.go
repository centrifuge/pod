package documents

import (
	"math/big"

	"github.com/centrifuge/go-centrifuge/errors"
)

// Int256 represents a signed 256 bit integer
type Int256 struct {
	v *big.Int
}

// NewInt256 creates a new Int256 given a string
func NewInt256(n string) (*Int256, error) {
	nn := new(big.Int)
	nn.SetString(n, 10)
	if !isValidInt256(*nn) {
		return nil, errors.NewTypedError(ErrInvalidInt256, errors.New("value: %s", n))
	}
	return &Int256{nn}, nil
}

// Bytes returns the byte representation of this int256
func (i *Int256) Bytes() []byte {
	return i.v.Bytes()
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
	if n.Abs(&n).Cmp(maxI256) > 0 {
		return false
	}

	return true
}
