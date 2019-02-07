// +build integration

package nft_test

import (
	"os"
	"testing"
	"time"

	"github.com/centrifuge/centrifuge-protobufs/documenttypes"
	"github.com/centrifuge/go-centrifuge/bootstrap"
	cc "github.com/centrifuge/go-centrifuge/bootstrap/bootstrappers/testingbootstrap"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/coredocument"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/documents/invoice"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/nft"
	"github.com/centrifuge/go-centrifuge/protobufs/gen/go/invoice"
	"github.com/centrifuge/go-centrifuge/testingutils/config"
	"github.com/centrifuge/go-centrifuge/testingutils/identity"
	"github.com/centrifuge/go-centrifuge/transactions"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
)

var registry *documents.ServiceRegistry
var cfg config.Configuration
var idService identity.Service
var payOb nft.PaymentObligation
var txManager transactions.Manager
var tokenRegistry coredocument.TokenRegistry

func TestMain(m *testing.M) {
	log.Debug("Test PreSetup for NFT")
	ctx := cc.TestFunctionalEthereumBootstrap()
	registry = ctx[documents.BootstrappedRegistry].(*documents.ServiceRegistry)
	idService = ctx[identity.BootstrappedIDService].(identity.Service)
	cfg = ctx[bootstrap.BootstrappedConfig].(config.Configuration)
	payOb = ctx[nft.BootstrappedPayObService].(nft.PaymentObligation)
	txManager = ctx[transactions.BootstrappedService].(transactions.Manager)
	tokenRegistry = ctx[nft.BootstrappedPayObService].(coredocument.TokenRegistry)
	result := m.Run()
	cc.TestFunctionalEthereumTearDown()
	os.Exit(result)
}

func TestPaymentObligationService_mint(t *testing.T) {
	// create identity
	log.Debug("Create Identity for Testing")
	cid := testingidentity.CreateIdentityWithKeys(cfg, idService)

	// create invoice (anchor)
	service, err := registry.LocateService(documenttypes.InvoiceDataTypeUrl)
	assert.Nil(t, err, "should not error out when getting invoice service")
	ctx := testingconfig.CreateAccountContext(t, cfg)
	invoiceService := service.(invoice.Service)
	dueDate := time.Now().Add(4 * 24 * time.Hour)
	model, err := invoiceService.DeriveFromCreatePayload(ctx, &invoicepb.InvoiceCreatePayload{
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
	modelUpdated, txID, _, err := invoiceService.Create(ctx, model)
	err = txManager.WaitForTransaction(cid, txID)
	assert.Nil(t, err)

	// get ID
	ID, err := modelUpdated.ID()
	assert.Nil(t, err, "should not error out when getting invoice ID")
	// call mint
	// assert no error
	depositAddr := "0xf72855759a39fb75fc7341139f5d7a3974d4da08"
	registry := cfg.GetContractAddress(config.PaymentObligation)
	resp, done, err := payOb.MintNFT(
		ctx,
		ID,
		registry.String(),
		depositAddr,
		[]string{"invoice.gross_amount", "invoice.currency", "invoice.due_date", "collaborators[0]"},
	)
	assert.Nil(t, err, "should not error out when minting an invoice")
	assert.NotNil(t, resp.TokenID, "token id should be present")
	tokenID, err := nft.FromString(resp.TokenID)
	assert.Nil(t, err, "should not error out when getting tokenID hex")
	<-done
	assert.NoError(t, txManager.WaitForTransaction(cid, uuid.Must(uuid.FromString(resp.TransactionID))))
	owner, err := tokenRegistry.OwnerOf(registry, tokenID.BigInt().Bytes())
	assert.NoError(t, err)
	assert.Equal(t, common.HexToAddress(depositAddr), owner)
	doc, err := invoiceService.GetCurrentVersion(ctx, ID)
	assert.NoError(t, err)
	cd, err := doc.PackCoreDocument()
	assert.NoError(t, err)
	assert.Len(t, cd.Roles, 2)
	assert.Len(t, cd.Roles[1].Nfts, 1)
	newNFT := cd.Roles[1].Nfts[0]
	enft, err := coredocument.ConstructNFT(registry, tokenID.BigInt().Bytes())
	assert.NoError(t, err)
	assert.Equal(t, enft, newNFT)

	// try to mint the NFT again
	_, _, err = payOb.MintNFT(ctx,
		ID,
		registry.String(),
		depositAddr,
		[]string{"invoice.gross_amount", "invoice.currency", "invoice.due_date", "collaborators[0]"})
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(nft.ErrNFTMinted, err))
}
