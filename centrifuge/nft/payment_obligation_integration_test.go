// +build integration

package nft_test

import (
	"os"
	"testing"

	"context"
	"time"

	"github.com/centrifuge/centrifuge-protobufs/documenttypes"
	cc "github.com/centrifuge/go-centrifuge/centrifuge/context/testingbootstrap"
	"github.com/centrifuge/go-centrifuge/centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/centrifuge/documents/invoice"
	"github.com/centrifuge/go-centrifuge/centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/centrifuge/nft"
	"github.com/centrifuge/go-centrifuge/centrifuge/protobufs/gen/go/invoice"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/stretchr/testify/assert"
)

var identityService identity.Service

func TestMain(m *testing.M) {
	cc.TestFunctionalEthereumBootstrap()
	identityService = identity.IDService
	result := m.Run()
	cc.TestFunctionalEthereumTearDown()
	os.Exit(result)
}

func TestPaymentObligationService_mint(t *testing.T) {
	// create identity
	centID := identity.NewRandomCentID()
	createIdentityWithKeys(t, centID[:])

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
	// TODO setup the payob contract during integration test init
	_, err = nft.GetPaymentObligationService().MintNFT(
		ID,
		documenttypes.InvoiceDataTypeUrl,
		"doesntmatter",
		"doesntmatter",
		[]string{"gross_amount", "currency", "due_date"},
	)
	assert.Nil(t, err, "should not error out when minting an invoice")
}

func createIdentityWithKeys(t *testing.T, centrifugeId []byte) []byte {

	centIdTyped, _ := identity.ToCentID(centrifugeId)
	id, confirmations, err := identityService.CreateIdentity(centIdTyped)
	assert.Nil(t, err, "should not error out when creating identity")

	watchRegisteredIdentity := <-confirmations
	assert.Nil(t, watchRegisteredIdentity.Error, "No error thrown by context")

	// LookupIdentityForId
	id, err = identityService.LookupIdentityForID(centIdTyped)
	assert.Nil(t, err, "should not error out when resolving identity")

	pubKey, _ := hexutil.Decode("0xc8dd3d66e112fae5c88fe6a677be24013e53c33e")

	confirmations, err = id.AddKeyToIdentity(context.Background(), identity.KeyPurposeEthMsgAuth, pubKey)
	assert.Nil(t, err, "should not error out when adding keys")
	assert.NotNil(t, confirmations, "confirmations channel should not be nil")
	watchRegisteredIdentityKey := <-confirmations
	assert.Nil(t, watchRegisteredIdentityKey.Error, "No error thrown by context")

	return centrifugeId
}
