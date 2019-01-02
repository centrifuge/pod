// +build integration

package nft_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/centrifuge/centrifuge-protobufs/documenttypes"
	"github.com/centrifuge/go-centrifuge/bootstrap"
	cc "github.com/centrifuge/go-centrifuge/bootstrap/bootstrappers/testingbootstrap"
	ccommon "github.com/centrifuge/go-centrifuge/common"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/contextutil"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/documents/invoice"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/nft"
	"github.com/centrifuge/go-centrifuge/protobufs/gen/go/invoice"
	"github.com/centrifuge/go-centrifuge/testingutils/identity"
	"github.com/centrifuge/go-centrifuge/transactions"
	"github.com/ethereum/go-ethereum/log"
	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
)

var registry *documents.ServiceRegistry
var cfg config.Configuration
var idService identity.Service
var payOb nft.PaymentObligation
var txService transactions.Service

func TestMain(m *testing.M) {
	log.Debug("Test PreSetup for NFT")
	ctx := cc.TestFunctionalEthereumBootstrap()
	registry = ctx[documents.BootstrappedRegistry].(*documents.ServiceRegistry)
	idService = ctx[identity.BootstrappedIDService].(identity.Service)
	cfg = ctx[bootstrap.BootstrappedConfig].(config.Configuration)
	payOb = ctx[nft.BootstrappedPayObService].(nft.PaymentObligation)
	txService = ctx[transactions.BootstrappedService].(transactions.Service)
	result := m.Run()
	cc.TestFunctionalEthereumTearDown()
	os.Exit(result)
}

func TestPaymentObligationService_mint(t *testing.T) {
	// create identity
	log.Debug("Create Identity for Testing")
	testingidentity.CreateIdentityWithKeys(cfg, idService)

	// create invoice (anchor)
	service, err := registry.LocateService(documenttypes.InvoiceDataTypeUrl)
	assert.Nil(t, err, "should not error out when getting invoice service")
	contextHeader, err := contextutil.NewCentrifugeContext(context.Background(), cfg)
	assert.Nil(t, err)
	invoiceService := service.(invoice.Service)
	dueDate := time.Now().Add(4 * 24 * time.Hour)
	model, err := invoiceService.DeriveFromCreatePayload(contextHeader, &invoicepb.InvoiceCreatePayload{
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
	modelUpdated, _, err := invoiceService.Create(contextHeader, model)

	// get ID
	ID, err := modelUpdated.ID()
	assert.Nil(t, err, "should not error out when getting invoice ID")
	// call mint
	// assert no error
	resp, err := payOb.MintNFT(
		ccommon.DummyIdentity,
		ID,
		cfg.GetContractAddress("paymentObligation").String(),
		"0xf72855759a39fb75fc7341139f5d7a3974d4da08",
		[]string{"invoice.gross_amount", "invoice.currency", "invoice.due_date", "collaborators[0]"},
	)
	assert.Nil(t, err, "should not error out when minting an invoice")

	confirm, err := waitTillSuccessOrError(resp.TransactionID, txService)
	assert.True(t, confirm)
	assert.Nil(t, err)
	assert.NotNil(t, resp.TokenID, "token id should be present")
}

func waitTillSuccessOrError(id string, txService transactions.Service) (bool, error) {
	txID := uuid.Must(uuid.FromString(id))
	status := transactions.Pending
	for status == transactions.Pending {
		resp, err := txService.GetTransactionStatus(ccommon.DummyIdentity, txID)
		if err != nil {
			return false, err
		}

		if resp.Status == string(transactions.Pending) {
			continue
		}

		status = transactions.Status(resp.Status)
		break
	}

	return status == transactions.Success, nil
}
