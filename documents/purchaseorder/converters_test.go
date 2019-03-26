// +build unit

package purchaseorder

import (
	"testing"

	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/stretchr/testify/assert"
)

func TestLineItemActivities(t *testing.T) {
	dec := new(documents.Decimal)
	assert.NoError(t, dec.SetString("1.1"))
	items := []*LineItemActivity{
		{
			ItemNumber: "12345",
			Status:     "pending",
			Amount:     dec,
		},
	}

	citems := toClientActivities(items)
	pitems, err := toP2PActivities(items)
	assert.NoError(t, err)

	fcitems, err := fromClientLineItemActivities(citems)
	assert.NoError(t, err)
	assert.Equal(t, items, fcitems)

	fpitems, err := fromP2PLineItemActivities(pitems)
	assert.NoError(t, err)
	assert.Equal(t, items, fpitems)

	citems[0].Amount = "0.1.1"
	_, err = fromClientLineItemActivities(citems)
	assert.Error(t, err)

	pitems[0].Amount = utils.RandomSlice(40)
	_, err = fromP2PLineItemActivities(pitems)
	assert.Error(t, err)
}

func TestTaxItems(t *testing.T) {
	dec := new(documents.Decimal)
	assert.NoError(t, dec.SetString("1.1"))
	items := []*TaxItem{
		{
			ItemNumber: "12345",
			TaxAmount:  dec,
		},
	}

	citems := toClientTaxItems(items)
	pitems, err := toP2PTaxItems(items)
	assert.NoError(t, err)

	fcitems, err := fromClientTaxItems(citems)
	assert.NoError(t, err)
	assert.Equal(t, items, fcitems)

	fpitems, err := fromP2PTaxItems(pitems)
	assert.NoError(t, err)
	assert.Equal(t, items, fpitems)

	citems[0].TaxAmount = "1.1.1"
	_, err = fromClientTaxItems(citems)
	assert.Error(t, err)

	pitems[0].TaxAmount = utils.RandomSlice(40)
	_, err = fromP2PTaxItems(pitems)
	assert.Error(t, err)
}

func TestLineItems(t *testing.T) {
	dec := new(documents.Decimal)
	assert.NoError(t, dec.SetString("1.1"))
	items := []*LineItem{
		{
			Status:      "pending",
			AmountTotal: dec,
			Activities: []*LineItemActivity{
				{
					ItemNumber: "12345",
					Status:     "pending",
					Amount:     dec,
				},
			},
			TaxItems: []*TaxItem{
				{
					ItemNumber: "12345",
					TaxAmount:  dec,
				},
			},
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

	citems[0].Activities[0].Amount = "0.1.1"
	_, err = fromClientLineItems(citems)
	assert.Error(t, err)

	citems[0].TaxItems[0].TaxAmount = "1.1.1"
	_, err = fromClientLineItems(citems)
	assert.Error(t, err)

	citems[0].AmountTotal = "0.1.1"
	_, err = fromClientLineItems(citems)
	assert.Error(t, err)

	rdec := utils.RandomSlice(40)
	pitems[0].Activities[0].Amount = rdec
	_, err = fromP2PLineItems(pitems)
	assert.Error(t, err)

	pitems[0].TaxItems[0].TaxAmount = rdec
	_, err = fromP2PLineItems(pitems)
	assert.Error(t, err)

	pitems[0].AmountTotal = rdec
	_, err = fromP2PLineItems(pitems)
	assert.Error(t, err)
}
