// +build integration

package nft_test

import (
	"os"
	"testing"

	"context"
	"time"

	"github.com/centrifuge/centrifuge-protobufs/documenttypes"
	"github.com/centrifuge/go-centrifuge/centrifuge/config"
	cc "github.com/centrifuge/go-centrifuge/centrifuge/context/testingbootstrap"
	"github.com/centrifuge/go-centrifuge/centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/centrifuge/documents/invoice"
	"github.com/centrifuge/go-centrifuge/centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/centrifuge/nft"
	"github.com/centrifuge/go-centrifuge/centrifuge/protobufs/gen/go/invoice"
	"github.com/centrifuge/go-centrifuge/centrifuge/testingutils"
	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/stretchr/testify/assert"
)

var identityService identity.Service

func TestMain(m *testing.M) {
	cc.DONT_USE_FOR_UNIT_TESTS_TestFunctionalEthereumBootstrap()
	identityService = identity.IDService
	config.Config.V.Set("keys.signing.publicKey", "../../example/resources/signature1.pub.pem")
	config.Config.V.Set("keys.signing.privateKey", "../../example/resources/signature1.key.pem")
	config.Config.V.Set("keys.ethauth.publicKey", "../../example/resources/ethauth.pub.pem")
	config.Config.V.Set("keys.ethauth.privateKey", "../../example/resources/ethauth.key.pem")
	result := m.Run()
	cc.TestFunctionalEthereumTearDown()
	os.Exit(result)
}

func TestPaymentObligationService_mint(t *testing.T) {
	// create identity
	testingutils.CreateIdentityWithKeys()

	// create invoice (anchor)
	service, err := documents.GetRegistryInstance().LocateService(documenttypes.InvoiceDataTypeUrl)
	assert.Nil(t, err, "should not error out when getting invoice service")

	invoiceService := service.(invoice.Service)
	dueDate := time.Now().Add(4 * 24 * time.Hour)
	model, err := invoiceService.DeriveFromCreatePayload(
		&invoicepb.InvoiceCreatePayload{
			Collaborators: []string{},
			Data: &invoicepb.InvoiceData{
				InvoiceNumber: "2132131",
				GrossAmount:   123,
				NetAmount:     123,
				Currency:      "EUR",
				DueDate:       &timestamp.Timestamp{Seconds: dueDate.Unix()},
			},
		})
	assert.Nil(t, err, "should not error out when creating invoice model")
	modelUpdated, err := invoiceService.Create(context.Background(), model)

	// get ID
	ID, err := modelUpdated.GetDocumentID()
	assert.Nil(t, err, "should not error out when getting invoice ID")
	// call mint
	// assert no error
	_, err = nft.GetPaymentObligation().MintNFT(
		ID,
		documenttypes.InvoiceDataTypeUrl,
		"doesntmatter",
		"doesntmatter",
		[]string{"gross_amount", "currency", "due_date"},
	)
	assert.Nil(t, err, "should not error out when minting an invoice")
}
