//go:build unit

package documents

import (
	"testing"

	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestDecimal(t *testing.T) {
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

		{
			s:   "0.09",
			res: "0.09",
		},

		{
			s:   "-0.09",
			res: "-0.09",
		},

		// min decimal
		{
			s:   "0.000000000000000001",
			res: "0.000000000000000001",
		},

		{
			s:   "-0.000000000000000001",
			res: "-0.000000000000000001",
		},

		{
			s:   "0.0000000000000001",
			res: "0.0000000000000001",
		},

		{
			s:   "-0.0000000000000001",
			res: "-0.0000000000000001",
		},

		// more than supported precision decimal
		{
			s:   "1.1234567891234567891",
			err: true,
		},

		// zero
		{
			s:   "0",
			res: "0",
		},

		// zero
		{
			s:   "0.0",
			res: "0",
		},

		// zero
		{
			s:   "-0",
			res: "0",
		},

		// zero
		{
			s:   "-0.0",
			res: "0",
		},

		// max int
		{
			s:   "999999999999999999999999999999999999999999999999999999",
			res: "999999999999999999999999999999999999999999999999999999",
		},

		// max neg int
		{
			s:   "-999999999999999999999999999999999999999999999999999999",
			res: "-999999999999999999999999999999999999999999999999999999",
		},

		// min int
		{
			s:   "1",
			res: "1",
		},

		// min neg int
		{
			s:   "-1",
			res: "-1",
		},

		// minimum fraction
		{
			s:   "0.1",
			res: "0.1",
		},

		// minimum neg fraction
		{
			s:   "-0.1",
			res: "-0.1",
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

		// max integer with min fraction
		{
			s:   "9999999999999999999999999999999999999999999999999999999.000000000000000001",
			res: "9999999999999999999999999999999999999999999999999999999.000000000000000001",
		},

		// max integer with min fraction
		{
			s:   "-9999999999999999999999999999999999999999999999999999999.000000000000000001",
			res: "-9999999999999999999999999999999999999999999999999999999.000000000000000001",
		},
	}

	for _, c := range tests {
		d := new(Decimal)
		err := d.SetString(c.s)
		if c.err {
			assert.Error(t, err)
			continue
		}

		assert.NoError(t, err)
		assert.Equal(t, c.res, d.String())

		b, err := d.Bytes()
		assert.NoError(t, err)

		d1 := new(Decimal)
		assert.NoError(t, d1.SetBytes(b))
		assert.Equal(t, c.res, d1.String())

		buf, err := d.MarshalJSON()
		assert.NoError(t, err)

		d2 := new(Decimal)
		assert.NoError(t, d2.UnmarshalJSON(buf))

		assert.Equal(t, c.res, d2.String())
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
	err := d.SetBytes(buf)
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(ErrInvalidDecimal, err))
}

func TestDecimalsToStrings(t *testing.T) {
	tests := []struct {
		s, res string
	}{
		{},

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
		std, err := StringsToDecimals(c.s)
		assert.NoError(t, err)

		dts := DecimalsToStrings(std...)
		assert.Len(t, dts, 1)
		assert.Equal(t, dts[0], c.res)

		bytes, err := DecimalsToBytes(std...)
		assert.NoError(t, err)

		btd, err := BytesToDecimals(bytes...)
		assert.NoError(t, err)
		assert.Len(t, btd, 1)
		var res string
		if btd[0] != nil {
			res = btd[0].String()
		}
		assert.Equal(t, res, c.res)
	}
}
