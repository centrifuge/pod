// +build unit

package documents

import (
	"testing"

	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestDecimal_FromString(t *testing.T) {
	tests := []struct {
		s, res string
		err    bool
	}{

		// empty
		{
			err: true,
		},

		// invalid
		{
			s:   "some invalid",
			err: true,
		},

		// invalid decimal
		{
			s:   "1.1234567891234567891",
			err: true,
		},

		// more than int
		{
			s:   "999999999999999999999999999999999999999999999999999999999",
			err: true,
		},

		// less than neg int
		{
			s:   "-999999999999999999999999999999999999999999999999999999999",
			err: true,
		},

		// int
		{
			s:   "999999999999999999999999999999999999999999999999999999",
			res: "999999999999999999999999999999999999999999999999999999",
		},

		// neg int
		{
			s:   "-999999999999999999999999999999999999999999999999999999",
			res: "-999999999999999999999999999999999999999999999999999999",
		},

		{
			s:   ".999999999999999999",
			res: "0.999999999999999999",
		},

		{
			s:   "-.999999999999999999",
			res: "-0.999999999999999999",
		},

		{
			s:   "999999999999999999.0",
			res: "999999999999999999",
		},

		{
			s:   "-999999999999999999.0",
			res: "-999999999999999999",
		},

		// decimal
		{
			s:   "9999999999999999999999999999999999999999999999999999999.999999999999999999",
			res: "9999999999999999999999999999999999999999999999999999999.999999999999999999",
		},

		// neg decimal
		{
			s:   "-9999999999999999999999999999999999999999999999999999999.999999999999999999",
			res: "-9999999999999999999999999999999999999999999999999999999.999999999999999999",
		},
	}

	for _, c := range tests {
		d := new(Decimal)
		err := d.FromString(c.s)
		if c.err {
			assert.Error(t, err)
			continue
		}

		assert.NoError(t, err)
		assert.Equal(t, c.res, d.String())

		b, err := d.Bytes()
		assert.NoError(t, err)

		d1 := new(Decimal)
		assert.NoError(t, d1.FromBytes(b))
		assert.Equal(t, c.res, d1.String())
	}
}

func TestDecimal_Bytes_max_byte_exceeded(t *testing.T) {
	dec, err := decimal.NewFromString("999999999999999999999999999999999999999999999999999999999")
	assert.NoError(t, err)
	d := new(Decimal)
	d.dec = dec
	buf, err := d.Bytes()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "max byte length exceeded")
	assert.Nil(t, buf)
}

func TestDecimal_FromBytes_min_byte_error(t *testing.T) {
	d := new(Decimal)
	buf := utils.RandomSlice(7)
	err := d.FromBytes(buf)
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(ErrInvalidDecimal, err))
}
