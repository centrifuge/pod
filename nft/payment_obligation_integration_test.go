// +build integration

package nft_test

import (
	"os"
	"testing"

	"github.com/centrifuge/go-centrifuge/config"
	cc "github.com/centrifuge/go-centrifuge/context/testingbootstrap"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/nft"
	"github.com/ethereum/go-ethereum/log"
)

var registry *documents.ServiceRegistry
var cfg *config.Configuration
var idService identity.Service
var payOb nft.PaymentObligation

func TestMain(m *testing.M) {
	log.Debug("Test PreSetup for NFT")
	ctx := cc.TestFunctionalEthereumBootstrap()
	registry = ctx[documents.BootstrappedRegistry].(*documents.ServiceRegistry)
	idService = ctx[identity.BootstrappedIDService].(identity.Service)
	cfg = ctx[config.BootstrappedConfig].(*config.Configuration)
	payOb = ctx[nft.BootstrappedPayObService].(nft.PaymentObligation)
	prevSignPubkey := cfg.Get("keys.signing.publicKey")
	prevSignPrivkey := cfg.Get("keys.signing.privateKey")
	prevEthPubkey := cfg.Get("keys.ethauth.publicKey")
	prevEthPrivkey := cfg.Get("keys.ethauth.privateKey")
	cfg.Set("keys.signing.publicKey", "../build/resources/signingKey.pub.pem")
	cfg.Set("keys.signing.privateKey", "../build/resources/signingKey.key.pem")
	cfg.Set("keys.ethauth.publicKey", "../build/resources/ethauth.pub.pem")
	cfg.Set("keys.ethauth.privateKey", "../build/resources/ethauth.key.pem")
	result := m.Run()
	cfg.Set("keys.signing.publicKey", prevSignPubkey)
	cfg.Set("keys.signing.privateKey", prevSignPrivkey)
	cfg.Set("keys.ethauth.publicKey", prevEthPubkey)
	cfg.Set("keys.ethauth.privateKey", prevEthPrivkey)
	cc.TestFunctionalEthereumTearDown()
	os.Exit(result)
}

//func TestPaymentObligationService_mint(t *testing.T) {
//	// create identity
//	log.Debug("Create Identity for Testing")
//	testingidentity.CreateIdentityWithKeys(cfg, idService)
//
//	// create invoice (anchor)
//	service, err := registry.LocateService(documenttypes.InvoiceDataTypeUrl)
//	assert.Nil(t, err, "should not error out when getting invoice service")
//	contextHeader, err := header.NewContextHeader(context.Background(), cfg)
//	assert.Nil(t, err)
//	invoiceService := service.(invoice.Service)
//	dueDate := time.Now().Add(4 * 24 * time.Hour)
//	model, err := invoiceService.DeriveFromCreatePayload(
//		&invoicepb.InvoiceCreatePayload{
//			Collaborators: []string{},
//			Data: &invoicepb.InvoiceData{
//				InvoiceNumber: "2132131",
//				GrossAmount:   123,
//				NetAmount:     123,
//				Currency:      "EUR",
//				DueDate:       &timestamp.Timestamp{Seconds: dueDate.Unix()},
//			},
//		}, contextHeader)
//	assert.Nil(t, err, "should not error out when creating invoice model")
//	modelUpdated, err := invoiceService.Create(contextHeader, model)
//
//	// get ID
//	ID, err := modelUpdated.ID()
//	assert.Nil(t, err, "should not error out when getting invoice ID")
//	// call mint
//	// assert no error
//	confirmations, err := payOb.MintNFT(
//		ID,
//		cfg.GetContractAddress("paymentObligation").String(),
//		"0xf72855759a39fb75fc7341139f5d7a3974d4da08",
//		[]string{"invoice.gross_amount", "invoice.currency", "invoice.due_date", "document_type", "collaborators[0]"},
//	)
//	assert.Nil(t, err, "should not error out when minting an invoice")
//	tokenConfirm := <-confirmations
//	assert.Nil(t, tokenConfirm.Err, "should not error out when minting an invoice")
//	assert.NotNil(t, tokenConfirm.TokenID, "token id should be present")
//}
