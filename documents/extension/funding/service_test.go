// +build unit

package funding

import (
	"github.com/centrifuge/go-centrifuge/documents/invoice"
	clientfundingpb "github.com/centrifuge/go-centrifuge/protobufs/gen/go/funding"
	"github.com/centrifuge/go-centrifuge/testingutils/documents"
	"github.com/centrifuge/go-centrifuge/testingutils/identity"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGenerateKey(t *testing.T) {
	assert.Equal(t,"centrifuge_funding[1].days",generateKey(1,"days"))
	assert.Equal(t,"centrifuge_funding[0].",generateKey(0,""))

}


func TestAddFundingAttributes(t *testing.T) {
	testingdocuments.CreateInvoicePayload()
	inv := &invoice.Invoice{}
	inv.InitInvoiceInput(testingdocuments.CreateInvoicePayload(),testingidentity.GenerateRandomDID())


	payload := &clientfundingpb.FundingCreatePayload{Data:&clientfundingpb.FundingData{Currency:"eur"}}

	attributes, err := createAttributeMap(inv,payload)
	assert.NoError(t, err)
	assert.Equal(t, "eur", attributes["centrifuge_funding[0].currency"])

}


