// +build integration

package nft_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/centrifuge/centrifuge-protobufs/documenttypes"
	"github.com/centrifuge/go-centrifuge/config"
	cc "github.com/centrifuge/go-centrifuge/context/testingbootstrap"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/documents/invoice"
	"github.com/centrifuge/go-centrifuge/nft"
	"github.com/centrifuge/go-centrifuge/protobufs/gen/go/invoice"
	"github.com/centrifuge/go-centrifuge/testingutils"
	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	cc.TestFunctionalEthereumBootstrap()
	prevSignPubkey := config.Config.V.Get("keys.signing.publicKey")
	prevSignPrivkey := config.Config.V.Get("keys.signing.privateKey")
	prevEthPubkey := config.Config.V.Get("keys.ethauth.publicKey")
	prevEthPrivkey := config.Config.V.Get("keys.ethauth.privateKey")
	config.Config.V.Set("keys.signing.publicKey", "../build/resources/signingKey.pub.pem")
	config.Config.V.Set("keys.signing.privateKey", "../build/resources/signingKey.key.pem")
	config.Config.V.Set("keys.ethauth.publicKey", "../build/resources/ethauth.pub.pem")
	config.Config.V.Set("keys.ethauth.privateKey", "../build/resources/ethauth.key.pem")
	result := m.Run()
	config.Config.V.Set("keys.signing.publicKey", prevSignPubkey)
	config.Config.V.Set("keys.signing.privateKey", prevSignPrivkey)
	config.Config.V.Set("keys.ethauth.publicKey", prevEthPubkey)
	config.Config.V.Set("keys.ethauth.privateKey", prevEthPrivkey)
	cc.TestFunctionalEthereumTearDown()
	os.Exit(result)
}

func TestPaymentObligationService_mint(t *testing.T) {
	// create identity
	testingutils.CreateIdentityWithKeys()

	// create invoice (anchor)
	service, err := documents.GetRegistryInstance().LocateService(documenttypes.InvoiceDataTypeUrl)
	assert.Nil(t, err, "should not error out when getting invoice service")
	contextHeader, err := documents.NewContextHeader()
	assert.Nil(t, err)
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
		}, contextHeader)
	assert.Nil(t, err, "should not error out when creating invoice model")
	modelUpdated, err := invoiceService.Create(context.Background(), model)

	// get ID
	ID, err := modelUpdated.ID()
	assert.Nil(t, err, "should not error out when getting invoice ID")
	// call mint
	// assert no error
	confirmations, err := nft.GetPaymentObligation().MintNFT(
		ID,
		documenttypes.InvoiceDataTypeUrl,
		"doesntmatter",
		"doesntmatter",
		[]string{"gross_amount", "currency", "due_date","document_type"},

	)
	assert.Nil(t, err, "should not error out when minting an invoice")
	tokenConfirm := <-confirmations
	assert.Nil(t, tokenConfirm.Err, "should not error out when minting an invoice")
	assert.NotNil(t, tokenConfirm.TokenID, "token id should be present")
}
