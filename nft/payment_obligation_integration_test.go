// +build integration

package nft_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/centrifuge/go-centrifuge/utils"

	"github.com/ethereum/go-ethereum/common/hexutil"

	"github.com/centrifuge/centrifuge-protobufs/documenttypes"
	"github.com/centrifuge/go-centrifuge/bootstrap"
	cc "github.com/centrifuge/go-centrifuge/bootstrap/bootstrappers/testingbootstrap"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/config/configstore"
	"github.com/centrifuge/go-centrifuge/contextutil"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/documents/invoice"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/nft"
	"github.com/centrifuge/go-centrifuge/protobufs/gen/go/invoice"
	"github.com/centrifuge/go-centrifuge/testingutils/identity"
	"github.com/centrifuge/go-centrifuge/transactions"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
	"github.com/stretchr/testify/assert"
)

var registry *documents.ServiceRegistry
var cfg config.Configuration
var cfgService config.Service
var idService identity.ServiceDID
var idFactory identity.Factory
var payOb nft.PaymentObligation
var txManager transactions.Manager
var tokenRegistry documents.TokenRegistry

func TestMain(m *testing.M) {
	log.Debug("Test PreSetup for NFT")
	ctx := cc.TestFunctionalEthereumBootstrap()
	registry = ctx[documents.BootstrappedRegistry].(*documents.ServiceRegistry)
	idService = ctx[identity.BootstrappedDIDService].(identity.ServiceDID)
	idFactory = ctx[identity.BootstrappedDIDFactory].(identity.Factory)
	cfg = ctx[bootstrap.BootstrappedConfig].(config.Configuration)
	cfgService = ctx[config.BootstrappedConfigStorage].(config.Service)
	payOb = ctx[nft.BootstrappedPayObService].(nft.PaymentObligation)
	txManager = ctx[transactions.BootstrappedService].(transactions.Manager)
	tokenRegistry = ctx[nft.BootstrappedPayObService].(documents.TokenRegistry)
	result := m.Run()
	cc.TestFunctionalEthereumTearDown()
	os.Exit(result)
}

func prepareForNFTMinting(t *testing.T) (context.Context, []byte, common.Address, string, documents.Service, identity.DID) {
	// create identity
	log.Debug("Create Identity for Testing")
	didAddr, err := idFactory.CalculateIdentityAddress(context.Background())
	assert.NoError(t, err)
	did := identity.NewDID(*didAddr)
	tc, err := configstore.TempAccount("main", cfg)
	assert.NoError(t, err)
	tcr := tc.(*configstore.Account)
	tcr.IdentityID = did[:]
	_, err = cfgService.CreateAccount(tcr)
	assert.NoError(t, err)
	cid, err := testingidentity.CreateAccountIDWithKeys(cfg.GetEthereumContextWaitTimeout(), tcr, idService, idFactory)
	assert.NoError(t, err)

	// create invoice (anchor)
	service, err := registry.LocateService(documenttypes.InvoiceDataTypeUrl)
	assert.Nil(t, err, "should not error out when getting invoice service")
	ctx, err := contextutil.New(context.Background(), tcr)
	assert.NoError(t, err)
	invSrv := service.(invoice.Service)
	dueDate := time.Now().Add(4 * 24 * time.Hour)
	tm, err := utils.ToTimestamp(dueDate)
	assert.NoError(t, err)
	model, err := invSrv.DeriveFromCreatePayload(ctx, &invoicepb.InvoiceCreatePayload{
		Collaborators: []string{},
		Data: &invoicepb.InvoiceData{
			Sender:      did.String(),
			Number:      "2132131",
			Status:      "unpaid",
			GrossAmount: "123",
			NetAmount:   "123",
			Currency:    "EUR",
			DateDue:     tm,
		},
	})
	assert.NoError(t, err, "should not error out when creating invoice model")
	modelUpdated, txID, done, err := invSrv.Create(ctx, model)
	assert.NoError(t, err)
	d := <-done
	assert.True(t, d)
	assert.NoError(t, txManager.WaitForTransaction(cid, txID))

	// get ID
	id := modelUpdated.ID()
	// call mint
	// assert no error
	depositAddr := "0xf72855759a39fb75fc7341139f5d7a3974d4da08"
	registry := cfg.GetContractAddress(config.PaymentObligation)

	return ctx, id, registry, depositAddr, invSrv, cid
}

func mintNFT(t *testing.T, ctx context.Context, req nft.MintNFTRequest, cid identity.DID, registry common.Address) nft.TokenID {
	resp, done, err := payOb.MintNFT(ctx, req)
	assert.NoError(t, err, "should not error out when minting an invoice")
	assert.NotNil(t, resp.TokenID, "token id should be present")
	tokenID, err := nft.TokenIDFromString(resp.TokenID)
	assert.NoError(t, err, "should not error out when getting tokenID hex")
	<-done
	txID, err := transactions.FromString(resp.TransactionID)
	assert.NoError(t, err)
	assert.NoError(t, txManager.WaitForTransaction(cid, txID))
	owner, err := tokenRegistry.OwnerOf(registry, tokenID.BigInt().Bytes())
	assert.NoError(t, err)
	assert.Equal(t, req.DepositAddress, owner)
	return tokenID
}

func TestPaymentObligationService_mint_grant_read_access(t *testing.T) {
	ctx, id, registry, depositAddr, invSrv, cid := prepareForNFTMinting(t)
	regAddr := registry.String()
	log.Info(regAddr)
	acc, err := contextutil.Account(ctx)
	assert.NoError(t, err)
	accDIDBytes, err := acc.GetIdentityID()
	assert.NoError(t, err)
	keys, err := acc.GetKeys()
	assert.NoError(t, err)
	signerId := hexutil.Encode(append(accDIDBytes, keys[identity.KeyPurposeSigning.Name].PublicKey...))
	signingRoot := fmt.Sprintf("%s.%s", documents.DRTreePrefix, documents.SigningRootField)
	signatureSender := fmt.Sprintf("%s.signatures[%s].signature", documents.SignaturesTreePrefix, signerId)
	req := nft.MintNFTRequest{
		DocumentID:               id,
		RegistryAddress:          registry,
		DepositAddress:           common.HexToAddress(depositAddr),
		ProofFields:              []string{"invoice.gross_amount", "invoice.currency", "invoice.date_due", "invoice.sender", "invoice.status", signingRoot, signatureSender, documents.CDTreePrefix + ".next_version"},
		GrantNFTReadAccess:       true,
		SubmitNFTReadAccessProof: true,
		SubmitTokenProof:         true,
	}
	tokenID := mintNFT(t, ctx, req, cid, registry)
	doc, err := invSrv.GetCurrentVersion(ctx, id)
	assert.NoError(t, err)
	cd, err := doc.PackCoreDocument()
	assert.NoError(t, err)
	assert.Len(t, cd.Roles, 3)
	assert.Len(t, cd.Roles[2].Nfts, 1)
	newNFT := cd.Roles[2].Nfts[0]
	enft, err := documents.ConstructNFT(registry, tokenID.BigInt().Bytes())
	assert.NoError(t, err)
	assert.Equal(t, enft, newNFT)

	// try to mint the NFT again
	_, _, err = payOb.MintNFT(ctx, req)
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(nft.ErrNFTMinted, err))
}

func failMintNFT(t *testing.T, grantNFT, nftReadAccess bool) {
	ctx, id, registry, depositAddr, _, _ := prepareForNFTMinting(t)
	req := nft.MintNFTRequest{
		DocumentID:               id,
		RegistryAddress:          registry,
		DepositAddress:           common.HexToAddress(depositAddr),
		ProofFields:              []string{"invoice.gross_amount", "invoice.currency", "invoice.due_date"},
		GrantNFTReadAccess:       grantNFT,
		SubmitNFTReadAccessProof: nftReadAccess,
	}

	_, _, err := payOb.MintNFT(ctx, req)
	assert.Error(t, err)
	if !nftReadAccess {
		assert.True(t, errors.IsOfType(documents.ErrNFTRoleMissing, err))
	}
}

func TestEthereumPaymentObligation_MintNFT_no_grant_access(t *testing.T) {
	failMintNFT(t, false, true)
}

func mintNFTWithProofs(t *testing.T, grantAccess, tokenProof, readAccessProof bool) {
	ctx, id, registry, depositAddr, invSrv, cid := prepareForNFTMinting(t)
	acc, err := contextutil.Account(ctx)
	assert.NoError(t, err)
	accDIDBytes, err := acc.GetIdentityID()
	assert.NoError(t, err)
	keys, err := acc.GetKeys()
	assert.NoError(t, err)
	signerId := hexutil.Encode(append(accDIDBytes, keys[identity.KeyPurposeSigning.Name].PublicKey...))
	signingRoot := fmt.Sprintf("%s.%s", documents.DRTreePrefix, documents.SigningRootField)
	signatureSender := fmt.Sprintf("%s.signatures[%s].signature", documents.SignaturesTreePrefix, signerId)
	req := nft.MintNFTRequest{
		DocumentID:               id,
		RegistryAddress:          registry,
		DepositAddress:           common.HexToAddress(depositAddr),
		ProofFields:              []string{"invoice.gross_amount", "invoice.currency", "invoice.due_date", "invoice.sender", "invoice.invoice_status", signingRoot, signatureSender, documents.CDTreePrefix + ".next_version"},
		GrantNFTReadAccess:       grantAccess,
		SubmitTokenProof:         tokenProof,
		SubmitNFTReadAccessProof: readAccessProof,
	}
	mintNFT(t, ctx, req, cid, registry)
	doc, err := invSrv.GetCurrentVersion(ctx, id)
	assert.NoError(t, err)
	cd, err := doc.PackCoreDocument()
	assert.NoError(t, err)
	roleCount := 2
	if grantAccess {
		roleCount++
	}
	assert.Len(t, cd.Roles, roleCount)
}

func TestEthereumPaymentObligation_MintNFT(t *testing.T) {
	tests := []struct {
		grantAccess, tokenProof, readAccessProof bool
	}{
		{
			grantAccess:     true,
			tokenProof:      true,
			readAccessProof: true,
		},
	}

	for _, c := range tests {
		mintNFTWithProofs(t, c.grantAccess, c.tokenProof, c.readAccessProof)
	}
}
