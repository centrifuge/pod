package documents

import (
	"fmt"
	"strings"

	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/utils/byteutils"
	"github.com/shopspring/decimal"
)

const (
	decimalPrecision      = 18
	maxIntegerLength      = 23
	maxFractionByteLength = 8
	minByteLength         = 9
	maxByteLength         = 32
)

// Decimal holds a fixed point decimal
type Decimal struct {
	dec decimal.Decimal
}

// FromString takes a decimal in string format and converts it into Decimal
func (d *Decimal) FromString(s string) error {
	s = strings.TrimSpace(s)
	if len(s) < 1 {
		return errors.New("empty string")
	}

	// normalise the string
	fd, err := decimal.NewFromString(s)
	if err != nil {
		return errors.NewTypedError(ErrInvalidDecimal, err)
	}

	s = fd.String()
	sd := strings.Split(s, ".")
	if len(sd) >= 2 && len(sd[1]) > decimalPrecision {
		return errors.New("exceeded max precision value 18: current %d", len(sd[1]))
	}

	integer, err := byteutils.IntBytesFromString(sd[0])
	if err != nil {
		return errors.NewTypedError(ErrInvalidDecimal, err)
	}

	if len(integer) > maxIntegerLength {
		return errors.NewTypedError(ErrInvalidDecimal, errors.New("integer exceeded max supported value"))
	}

	d.dec = fd
	return nil
}

// String returns the decimal in string representation.
func (d *Decimal) String() string {
	return d.dec.String()
}

// Bytes return the decimal in bytes.
// sign byte + upto 23 integer bytes + 8 decimal bytes
func (d *Decimal) Bytes() (decimal []byte, err error) {
	var sign byte
	if d.dec.Sign() < 0 {
		sign = byte(1)
	}
	decimal = []byte{sign}

	s := d.dec.Abs().String()
	sd := strings.Split(s, ".")

	fraction := make([]byte, maxFractionByteLength)
	if len(sd) >= 2 {
		fraction, err = byteutils.IntBytesFromString(sd[1])
		if err != nil {
			return nil, err
		}

		fraction = byteutils.AddZeroBytesSuffix(fraction, maxFractionByteLength)
	}

	integer, err := byteutils.IntBytesFromString(sd[0])
	if err != nil {
		return nil, err
	}

	decimal = append(decimal, integer...)
	decimal = append(decimal, fraction...)

	// sanity check
	// happens if we have done some calculations post conversion to Decimal.
	if len(decimal) > maxByteLength {
		return nil, errors.New("max byte length exceeded")
	}

	return decimal, nil
}

// FromBytes parse the bytes to Decimal.
func (d *Decimal) FromBytes(dec []byte) error {
	if len(dec) < minByteLength {
		return ErrInvalidDecimal
	}

	sign, dec := dec[0], dec[1:]
	fidx := len(dec) - maxFractionByteLength
	integer, fraction := byteutils.IntBytesToString(dec[:fidx]), byteutils.IntBytesToString(dec[fidx:])
	s := fmt.Sprintf("%s.%s", integer, fraction)
	if sign == 1 {
		s = "-" + s
	}

	return d.FromString(s)
}
