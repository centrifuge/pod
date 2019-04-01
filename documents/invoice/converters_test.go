// +build unit

package invoice

import (
	"testing"

	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/stretchr/testify/assert"
)

func TestLineItems(t *testing.T) {
	dec := new(documents.Decimal)
	err := dec.SetString("0.99")
	assert.NoError(t, err)
	items := []*LineItem{
		{
			ItemNumber: "123",
			TaxAmount:  dec,
		},
	}

	citems := toClientLineItems(items)
	pitems, err := toP2PLineItems(items)
	assert.NoError(t, err)

	fcitems, err := fromClientLineItems(citems)
	assert.NoError(t, err)
	assert.Equal(t, items, fcitems)

	fpitems, err := fromP2PLineItems(pitems)
	assert.NoError(t, err)
	assert.Equal(t, items, fpitems)

	citems[0].TaxAmount = "100.1234567891234567891"
	_, err = fromClientLineItems(citems)
	assert.Error(t, err)

	pitems[0].TaxAmount = utils.RandomSlice(40)
	_, err = fromP2PLineItems(pitems)
	assert.Error(t, err)
}

func TestTaxItems(t *testing.T) {
	dec := new(documents.Decimal)
	err := dec.SetString("0.99")
	assert.NoError(t, err)
	items := []*TaxItem{
		{
			ItemNumber: "some number",
			TaxAmount:  dec,
		},
	}

	citems := toClientTaxItems(items)
	pitems, err := toP2PTaxItems(items)
	assert.NoError(t, err)

	fcitems, err := fromClientTaxItems(citems)
	assert.NoError(t, err)
	fpitems, err := fromP2PTaxItems(pitems)
	assert.NoError(t, err)

	assert.Equal(t, items, fcitems)
	assert.Equal(t, items, fpitems)

	citems[0].TaxAmount = "0.1.1"
	_, err = fromClientTaxItems(citems)
	assert.Error(t, err)

	pitems[0].TaxAmount = utils.RandomSlice(40)
	_, err = fromP2PTaxItems(pitems)
	assert.Error(t, err)
}
