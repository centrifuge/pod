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

	pitems, err := toP2PLineItems(items)
	assert.NoError(t, err)

	fpitems, err := fromP2PLineItems(pitems)
	assert.NoError(t, err)
	assert.Equal(t, items, fpitems)

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

	pitems, err := toP2PTaxItems(items)
	assert.NoError(t, err)

	fpitems, err := fromP2PTaxItems(pitems)
	assert.NoError(t, err)
	assert.Equal(t, items, fpitems)

	pitems[0].TaxAmount = utils.RandomSlice(40)
	_, err = fromP2PTaxItems(pitems)
	assert.Error(t, err)
}
