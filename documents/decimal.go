package documents

import (
	"fmt"
	"math/big"
	"strings"

	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/utils/byteutils"
	"github.com/shopspring/decimal"
)

const (
	decimalPrecision     = 18
	maxIntegerByteLength = 23
	maxDecimalByteLength = 32
)

// Decimal holds a fixed point decimal
type Decimal struct {
	dec decimal.Decimal
}

// MarshalJSON marshals decimal to json bytes.
func (d *Decimal) MarshalJSON() ([]byte, error) {
	return d.dec.MarshalJSON()
}

// UnmarshalJSON loads json bytes to decimal
func (d *Decimal) UnmarshalJSON(data []byte) error {
	dec := new(decimal.Decimal)
	err := dec.UnmarshalJSON(data)
	if err != nil {
		return err
	}

	d.dec = *dec
	return nil
}

// SetString takes a decimal in string format and converts it into Decimal
func (d *Decimal) SetString(s string) error {
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
	if len(sd) == 2 && len(sd[1]) > decimalPrecision {
		return errors.New("exceeded max precision value 18: current %d", len(sd[1]))
	}

	integer, err := byteutils.IntBytesFromString(sd[0])
	if err != nil {
		return errors.NewTypedError(ErrInvalidDecimal, err)
	}

	if len(integer) > maxIntegerByteLength {
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
	if d.dec.IsZero() {
		return make([]byte, maxDecimalByteLength), nil
	}

	var sign byte
	if d.dec.Sign() < 0 {
		sign = byte(1)
	}
	decimal = []byte{sign}

	s := d.dec.Abs().String()
	sd := strings.Split(s, ".")

	v := sd[0]
	fraction := strings.Repeat("0", decimalPrecision)
	if len(sd) == 2 {
		// add extra zeros if precision is not max
		if cl := len(sd[1]); decimalPrecision-cl > 0 {
			sd[1] += strings.Repeat("0", decimalPrecision-cl)
		}

		fraction = sd[1]
	}

	v += fraction
	integer, err := byteutils.IntBytesFromString(v)
	if err != nil {
		return nil, err
	}

	// integer bytes should be atleast 32 byte, prepend Zeroes if len(integer)+sign(1) < maxDecimalByteLength
	if len(integer) < maxDecimalByteLength {
		integer = append(make([]byte, maxDecimalByteLength-(len(integer)+1)), integer...)
	}

	decimal = append(decimal, integer...)

	// sanity check
	// happens if we have done some calculations post conversion to Decimal.
	if len(decimal) > maxDecimalByteLength {
		return nil, errors.New("max byte length exceeded")
	}

	return decimal, nil
}

// SetBytes parse the bytes to Decimal.
func (d *Decimal) SetBytes(dec []byte) error {
	if len(dec) != maxDecimalByteLength {
		return ErrInvalidDecimal
	}

	sign, dec := dec[0], dec[1:]
	i := new(big.Int).SetBytes(dec)
	if i.Sign() == 0 {
		return d.SetString(i.String())
	}

	s := i.String()
	// edge case for only fractions
	// if s is less than 18 bytes, then its just a fraction, so convert to string and prepend required zeroes
	if decimalPrecision-len(s) > 0 {
		s = strings.Repeat("0", decimalPrecision-len(s)) + s
	}

	fidx := len(s) - decimalPrecision
	s = fmt.Sprintf("%s.%s", s[:fidx], s[fidx:])
	if sign == 1 {
		s = "-" + s
	}

	return d.SetString(s)
}

// NewDecimal returns a new decimal from given string
func NewDecimal(s string) (*Decimal, error) {
	dec := new(Decimal)
	err := dec.SetString(s)
	if err != nil {
		return nil, err
	}
	return dec, nil
}

// DecimalFromBytes returns a new decimal from bytes
func DecimalFromBytes(b []byte) (*Decimal, error) {
	d := new(Decimal)
	err := d.SetBytes(b)
	return d, err
}

// DecimalsToStrings converts decimals to string.
// nil decimal leads to empty string.
func DecimalsToStrings(decs ...*Decimal) []string {
	sdecs := make([]string, len(decs))
	for i, d := range decs {
		if d == nil {
			continue
		}

		sdecs[i] = d.String()
	}

	return sdecs
}

// DecimalsToBytes converts decimals to bytes
func DecimalsToBytes(decs ...*Decimal) ([][]byte, error) {
	dbytes := make([][]byte, len(decs))
	for i, d := range decs {
		if d == nil {
			continue
		}

		buf, err := d.Bytes()
		if err != nil {
			return nil, err
		}

		dbytes[i] = buf
	}

	return dbytes, nil
}

// StringsToDecimals converts string decimals to Decimal type
func StringsToDecimals(strs ...string) ([]*Decimal, error) {
	decs := make([]*Decimal, len(strs))
	for i, s := range strs {
		if strings.TrimSpace(s) == "" {
			continue
		}

		dec, err := NewDecimal(s)
		if err != nil {
			return nil, err
		}

		decs[i] = dec
	}

	return decs, nil
}

// BytesToDecimals converts decimals in bytes to Decimal type
func BytesToDecimals(bytes ...[]byte) ([]*Decimal, error) {
	decs := make([]*Decimal, len(bytes))
	for i, d := range bytes {
		d := d
		if len(d) < 1 {
			continue
		}

		dec := new(Decimal)
		err := dec.SetBytes(d)
		if err != nil {
			return nil, err
		}

		decs[i] = dec
	}

	return decs, nil
}
