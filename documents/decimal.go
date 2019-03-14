package documents

import (
	"strings"

	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/utils/byteutils"
	"github.com/shopspring/decimal"
)

const (
	decimalPrecision      = 18
	maxFractionByteLength = 8
)

type Decimal struct {
	integer, fraction []byte
	negative          bool
}

func (d *Decimal) FromString(s string) error {
	// normalise the string
	fd, err := decimal.NewFromString(s)
	if err != nil {
		return errors.NewTypedError(ErrInvalidDecimal, err)
	}

	s = fd.String()

	var neg bool
	if strings.HasPrefix(s, "-") {
		neg = true
		s = strings.TrimPrefix(s, "-")
	}
	sd := strings.Split(s, ".")

	var fraction, integer []byte
	if len(sd) == 2 {
		if len(sd[1]) > decimalPrecision {
			return errors.New("exceeded max precision value 18: current %d", len(sd))
		}

		// convert to bytes
		// append the bytes if needed
		fraction, err = byteutils.IntBytesFromString(sd[1])
		if err != nil {
			return errors.NewTypedError(ErrInvalidDecimal, err)
		}
	}

	integer, err = byteutils.IntBytesFromString(sd[0])
	if err != nil {
		return errors.NewTypedError(ErrInvalidDecimal, err)
	}

	d.integer = integer
	d.fraction = fraction
	d.negative = neg
	return nil
}
