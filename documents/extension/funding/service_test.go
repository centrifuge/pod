// +build unit

package funding

import (
	"context"
	"fmt"
	"testing"

	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/documents/invoice"
	clientfundingpb "github.com/centrifuge/go-centrifuge/protobufs/gen/go/funding"
	"github.com/centrifuge/go-centrifuge/testingutils/documents"
	"github.com/centrifuge/go-centrifuge/testingutils/identity"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGenerateKey(t *testing.T) {
	assert.Equal(t, "centrifuge_funding[1].days", generateKey("1", "days"))
	assert.Equal(t, "centrifuge_funding[0].", generateKey("0", ""))

}

func TestCreateAttributesList(t *testing.T) {
	testingdocuments.CreateInvoicePayload()
	inv := &invoice.Invoice{}
	inv.InitInvoiceInput(testingdocuments.CreateInvoicePayload(), testingidentity.GenerateRandomDID())

	payload := &clientfundingpb.FundingCreatePayload{Data: &clientfundingpb.FundingData{Currency: "eur"}}

	attributes, err := createAttributesList(inv, payload)
	assert.NoError(t, err)
	assert.Equal(t, 11, len(attributes))

	for _, attribute := range attributes {
		if attribute.KeyLabel == "centrifuge_funding[0].currency" {
			assert.Equal(t, "eur", attribute.Value.Str)
			break
		}
	}
}

func TestDeriveFromPayload(t *testing.T) {
	testingdocuments.CreateInvoicePayload()
	inv := &invoice.Invoice{}
	inv.InitInvoiceInput(testingdocuments.CreateInvoicePayload(), testingidentity.GenerateRandomDID())

	docSrv := &testingdocuments.MockService{}
	docSrv.On("GetCurrentVersion", mock.Anything, mock.Anything).Return(inv, nil)

	srv := DefaultService(docSrv, nil)
	payload := &clientfundingpb.FundingCreatePayload{Data: &clientfundingpb.FundingData{Currency: "eur"}}

	for i := 0; i < 10; i++ {
		model, err := srv.DeriveFromPayload(context.Background(), payload, utils.RandomSlice(32))
		assert.NoError(t, err)
		label := fmt.Sprintf("centrifuge_funding[%d].currency", i)
		key, err := documents.AttrKeyFromLabel(label)
		assert.NoError(t, err)

		attr, err := model.GetAttribute(key)
		assert.NoError(t, err)

		assert.Equal(t, "eur", attr.Value.Str)

	}

}
