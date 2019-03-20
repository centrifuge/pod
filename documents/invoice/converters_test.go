// +build unit

package invoice

import (
	"testing"

	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/testingutils/identity"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/stretchr/testify/assert"
)

func TestBinaryAttachments(t *testing.T) {
	atts := []*BinaryAttachment{
		{
			Name:     "some name",
			FileType: "pdf",
			Size:     1024,
			Data:     utils.RandomSlice(32),
			Checksum: utils.RandomSlice(32),
		},

		{
			Name:     "some name 1",
			FileType: "jpeg",
			Size:     4096,
			Data:     utils.RandomSlice(32),
			Checksum: utils.RandomSlice(32),
		},
	}

	catts := toClientAttachments(atts)
	patts := toP2PAttachments(atts)

	fcatts, err := fromClientAttachments(catts)
	assert.NoError(t, err)
	assert.Equal(t, atts, fcatts)

	fpatts := fromP2PAttachments(patts)
	assert.Equal(t, atts, fpatts)

	catts[0].Checksum = "some checksum"
	_, err = fromClientAttachments(catts)
	assert.Error(t, err)

	catts[0].Data = "some data"
	_, err = fromClientAttachments(catts)
	assert.Error(t, err)
}

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

func TestPaymentDetails(t *testing.T) {
	did := testingidentity.GenerateRandomDID()
	dec := new(documents.Decimal)
	err := dec.SetString("0.99")
	assert.NoError(t, err)
	details := []*PaymentDetails{
		{
			ID:     "some id",
			Payee:  &did,
			Amount: dec,
		},
	}

	cdetails := toClientPaymentDetails(details)
	pdetails, err := toP2PPaymentDetails(details)
	assert.NoError(t, err)

	fcdetails, err := fromClientPaymentDetails(cdetails)
	assert.NoError(t, err)
	fpdetails, err := fromP2PPaymentDetails(pdetails)
	assert.NoError(t, err)

	assert.Equal(t, details, fcdetails)
	assert.Equal(t, details, fpdetails)

	cdetails[0].Payee = "some did"
	_, err = fromClientPaymentDetails(cdetails)
	assert.Error(t, err)

	cdetails[0].Amount = "0.1.1"
	_, err = fromClientPaymentDetails(cdetails)
	assert.Error(t, err)

	pdetails[0].Amount = utils.RandomSlice(40)
	_, err = fromP2PPaymentDetails(pdetails)
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
