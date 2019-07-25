// +build integration

package nft_test

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/centrifuge/centrifuge-protobufs/documenttypes"
	"github.com/centrifuge/go-centrifuge/bootstrap"
	cc "github.com/centrifuge/go-centrifuge/bootstrap/bootstrappers/testingbootstrap"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/config/configstore"
	"github.com/centrifuge/go-centrifuge/contextutil"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/documents/generic"
	"github.com/centrifuge/go-centrifuge/documents/invoice"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/jobs"
	"github.com/centrifuge/go-centrifuge/nft"
	"github.com/centrifuge/go-centrifuge/testingutils"
	"github.com/centrifuge/go-centrifuge/testingutils/identity"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/log"
	"github.com/stretchr/testify/assert"
)

var registry *documents.ServiceRegistry
var cfg config.Configuration
var cfgService config.Service
var idService identity.Service
var idFactory identity.Factory
var invoiceUnpaid nft.Service
var jobManager jobs.Manager
var tokenRegistry documents.TokenRegistry

func TestMain(m *testing.M) {
	log.Debug("Test PreSetup for NFT")
	ctx := cc.TestFunctionalEthereumBootstrap()
	registry = ctx[documents.BootstrappedRegistry].(*documents.ServiceRegistry)
	idService = ctx[identity.BootstrappedDIDService].(identity.Service)
	idFactory = ctx[identity.BootstrappedDIDFactory].(identity.Factory)
	cfg = ctx[bootstrap.BootstrappedConfig].(config.Configuration)
	cfgService = ctx[config.BootstrappedConfigStorage].(config.Service)
	invoiceUnpaid = ctx[bootstrap.BootstrappedInvoiceUnpaid].(nft.Service)
	jobManager = ctx[jobs.BootstrappedService].(jobs.Manager)
	tokenRegistry = ctx[bootstrap.BootstrappedInvoiceUnpaid].(documents.TokenRegistry)
	result := m.Run()
	cc.TestFunctionalEthereumTearDown()
	os.Exit(result)
}

func prepareInvoiceForNFTMinting(t *testing.T) (context.Context, []byte, common.Address, documents.Service, identity.DID) {
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
	invSrv, err := registry.LocateService(documenttypes.InvoiceDataTypeUrl)
	assert.Nil(t, err, "should not error out when getting invoice invSrv")
	ctx, err := contextutil.New(context.Background(), tcr)
	assert.NoError(t, err)
	dueDate := time.Now().Add(4 * 24 * time.Hour)
	assert.NoError(t, err)

	payload := invoicePayload(t, nil, invoiceData(t, &did, &dueDate))
	model := invoice.InitInvoice(t, did, payload)
	assert.NoError(t, err, "should not error out when creating invoice model")
	modelUpdated, txID, done, err := invSrv.Create(ctx, model)
	assert.NoError(t, err)
	err = <-done
	assert.NoError(t, err)
	assert.NoError(t, jobManager.WaitForJob(cid, txID))

	// get ID
	id := modelUpdated.ID()
	// call mint
	// assert no error
	registry := cfg.GetContractAddress(config.InvoiceUnpaidNFT)

	return ctx, id, registry, invSrv, cid
}

func prepareGenericForNFTMinting(t *testing.T, regAddr string, attrs map[documents.AttrKey]documents.Attribute) (context.Context, []byte, common.Address, documents.Service, identity.DID) {
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

	// create Generic doc (anchor)
	genericSrv, err := registry.LocateService(documenttypes.GenericDataTypeUrl)
	assert.Nil(t, err, "should not error out when getting generic genSrv")
	ctx, err := contextutil.New(context.Background(), tcr)
	assert.NoError(t, err)

	payload := genericPayload(t, nil, attrs)
	modelUpdated, txID, err := genericSrv.CreateModel(ctx, payload)
	assert.NoError(t, err)
	assert.NoError(t, jobManager.WaitForJob(cid, txID))

	// get ID
	id := modelUpdated.ID()
	registry := common.HexToAddress(regAddr)

	return ctx, id, registry, genericSrv, cid
}

func mintNFT(t *testing.T, ctx context.Context, req nft.MintNFTRequest, cid identity.DID, registry common.Address) nft.TokenID {
	resp, done, err := invoiceUnpaid.MintNFT(ctx, req)
	assert.NoError(t, err, "should not error out when minting an invoice")
	assert.NotNil(t, resp.TokenID, "token id should be present")
	tokenID, err := nft.TokenIDFromString(resp.TokenID)
	assert.NoError(t, err, "should not error out when getting tokenID hex")
	err = <-done
	assert.NoError(t, err)
	jobID, err := jobs.FromString(resp.JobID)
	assert.NoError(t, err)
	assert.NoError(t, jobManager.WaitForJob(cid, jobID))
	owner, err := tokenRegistry.OwnerOf(registry, tokenID.BigInt().Bytes())
	assert.NoError(t, err)
	assert.Equal(t, req.DepositAddress, owner)
	return tokenID
}

func TestInvoiceUnpaidService_mint_grant_read_access(t *testing.T) {
	ctx, id, registry, invSrv, cid := prepareInvoiceForNFTMinting(t)
	regAddr := registry.String()
	log.Info(regAddr)
	acc, err := contextutil.Account(ctx)
	assert.NoError(t, err)
	accDIDBytes := acc.GetIdentityID()
	keys, err := acc.GetKeys()
	assert.NoError(t, err)
	signerId := hexutil.Encode(append(accDIDBytes, keys[identity.KeyPurposeSigning.Name].PublicKey...))
	signingRoot := fmt.Sprintf("%s.%s", documents.DRTreePrefix, documents.SigningRootField)
	signatureSender := fmt.Sprintf("%s.signatures[%s].signature", documents.SignaturesTreePrefix, signerId)
	req := nft.MintNFTRequest{
		DocumentID:               id,
		RegistryAddress:          registry,
		DepositAddress:           cid.ToAddress(),
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
	_, _, err = invoiceUnpaid.MintNFT(ctx, req)
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(nft.ErrNFTMinted, err))
}

func TestGenericMintNFT(t *testing.T) {
	attrs := map[documents.AttrKey]documents.Attribute{}
	loanAmount := "loanAmount"
	loanAmountValue := "100"
	attr0, err := documents.NewAttribute(loanAmount, documents.AttrDecimal, loanAmountValue)
	assert.NoError(t, err)
	attrs[attr0.Key] = attr0
	asIsValue := "asIsValue"
	asIsValueValue := "1000"
	attr1, err := documents.NewAttribute(asIsValue, documents.AttrDecimal, asIsValueValue)
	assert.NoError(t, err)
	attrs[attr1.Key] = attr1
	afterRehabValue := "afterRehabValue"
	afterRehabValueValue := "2000"
	attr2, err := documents.NewAttribute(afterRehabValue, documents.AttrDecimal, afterRehabValueValue)
	assert.NoError(t, err)
	attrs[attr2.Key] = attr2
	scAddrs := testingutils.GetDAppSmartContractAddresses()
	fmt.Println(scAddrs)
	ctx, id, registry, invSrv, cid := prepareGenericForNFTMinting(t, scAddrs["genericNFT"], attrs)
	fmt.Println("Generic NFT Registry", scAddrs["genericNFT"])

	attributeLoanAmount := fmt.Sprintf("%s.attributes[%s].byte_val", documents.CDTreePrefix, attr0.Key.String())
	attributeAsIsVal := fmt.Sprintf("%s.attributes[%s].byte_val", documents.CDTreePrefix, attr1.Key.String())
	attributeAfterRehabVal := fmt.Sprintf("%s.attributes[%s].byte_val", documents.CDTreePrefix, attr2.Key.String())
	proofFields := []string{attributeLoanAmount, attributeAsIsVal, attributeAfterRehabVal}

	req := nft.MintNFTRequest{
		DocumentID:               id,
		RegistryAddress:          registry,
		DepositAddress:           cid.ToAddress(),
		ProofFields:              proofFields,
		GrantNFTReadAccess:       false,
		SubmitNFTReadAccessProof: false,
		SubmitTokenProof:         true,
		UseGeneric:               true,
	}
	_ = mintNFT(t, ctx, req, cid, registry)
	doc, err := invSrv.GetCurrentVersion(ctx, id)
	assert.NoError(t, err)
	cd, err := doc.PackCoreDocument()
	assert.NoError(t, err)
	assert.Len(t, cd.Roles, 2)
}

func failMintNFT(t *testing.T, grantNFT, nftReadAccess bool) {
	ctx, id, registry, _, cid := prepareInvoiceForNFTMinting(t)
	req := nft.MintNFTRequest{
		DocumentID:               id,
		RegistryAddress:          registry,
		DepositAddress:           cid.ToAddress(),
		ProofFields:              []string{"invoice.gross_amount", "invoice.currency", "invoice.date_due"},
		GrantNFTReadAccess:       grantNFT,
		SubmitNFTReadAccessProof: nftReadAccess,
	}

	_, _, err := invoiceUnpaid.MintNFT(ctx, req)
	assert.Error(t, err)
	if !nftReadAccess {
		assert.True(t, errors.IsOfType(documents.ErrNFTRoleMissing, err))
	}
}

func TestEthereumInvoiceUnpaid_MintNFT_no_grant_access(t *testing.T) {
	failMintNFT(t, false, true)
}

func mintNFTWithProofs(t *testing.T, grantAccess, tokenProof, readAccessProof bool) (context.Context, nft.TokenID, identity.DID) {
	ctx, id, registry, invSrv, cid := prepareInvoiceForNFTMinting(t)
	acc, err := contextutil.Account(ctx)
	assert.NoError(t, err)
	accDIDBytes := acc.GetIdentityID()
	keys, err := acc.GetKeys()
	assert.NoError(t, err)
	signerId := hexutil.Encode(append(accDIDBytes, keys[identity.KeyPurposeSigning.Name].PublicKey...))
	signingRoot := fmt.Sprintf("%s.%s", documents.DRTreePrefix, documents.SigningRootField)
	signatureSender := fmt.Sprintf("%s.signatures[%s].signature", documents.SignaturesTreePrefix, signerId)
	req := nft.MintNFTRequest{
		DocumentID:               id,
		RegistryAddress:          registry,
		DepositAddress:           cid.ToAddress(),
		ProofFields:              []string{"invoice.gross_amount", "invoice.currency", "invoice.date_due", "invoice.sender", "invoice.status", signingRoot, signatureSender, documents.CDTreePrefix + ".next_version"},
		GrantNFTReadAccess:       grantAccess,
		SubmitTokenProof:         tokenProof,
		SubmitNFTReadAccessProof: readAccessProof,
	}
	tokenID := mintNFT(t, ctx, req, cid, registry)
	doc, err := invSrv.GetCurrentVersion(ctx, id)
	assert.NoError(t, err)
	cd, err := doc.PackCoreDocument()
	assert.NoError(t, err)
	roleCount := 2
	if grantAccess {
		roleCount++
	}
	assert.Len(t, cd.Roles, roleCount)
	return ctx, tokenID, cid
}

func TestEthereumInvoiceUnpaid_MintNFT(t *testing.T) {
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

func TestTransferNFT(t *testing.T) {
	addresses := testingutils.GetSmartContractAddresses()
	registry := common.HexToAddress(addresses.InvoiceUnpaidAddr)
	ctx, tokenID, did := mintNFTWithProofs(t, true, true, true)
	to := common.HexToAddress("0xde0B295669a9FD93d5F28D9Ec85E40f4cb697BAe")
	owner, err := invoiceUnpaid.OwnerOf(registry, tokenID[:])
	assert.NoError(t, err)
	assert.Equal(t, owner, did.ToAddress())

	// successful
	resp, done, err := invoiceUnpaid.TransferFrom(ctx, registry, to, tokenID)
	assert.NoError(t, err)
	err = <-done
	assert.NoError(t, err)
	jobID, err := jobs.FromString(resp.JobID)
	assert.NoError(t, err)
	assert.NoError(t, jobManager.WaitForJob(did, jobID))

	owner, err = invoiceUnpaid.OwnerOf(registry, tokenID[:])
	assert.NoError(t, err)
	assert.Equal(t, owner, to)

	// should fail not owner anymore
	secondTo := common.HexToAddress("0xFBb1b73C4f0BDa4f67dcA266ce6Ef42f520fBB98")
	resp, done, err = invoiceUnpaid.TransferFrom(ctx, registry, secondTo, tokenID)
	assert.NoError(t, err)
	err = <-done
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "from address is not the owner of tokenID")
	jobID, err = jobs.FromString(resp.JobID)
	assert.NoError(t, err)

	err = jobManager.WaitForJob(did, jobID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "from address is not the owner of tokenID")
}

func invoicePayload(t *testing.T, collaborators []identity.DID, data []byte) documents.CreatePayload {
	return documents.CreatePayload{
		Scheme: invoice.Scheme,
		Collaborators: documents.CollaboratorsAccess{
			ReadWriteCollaborators: collaborators,
		},
		Data: data,
	}
}

func genericPayload(t *testing.T, collaborators []identity.DID, attrs map[documents.AttrKey]documents.Attribute) documents.CreatePayload {
	return documents.CreatePayload{
		Scheme: generic.Scheme,
		Collaborators: documents.CollaboratorsAccess{
			ReadWriteCollaborators: collaborators,
		},
		Attributes: attrs,
	}
}

func invoiceData(t *testing.T, did *identity.DID, tm *time.Time) []byte {
	dec, err := documents.NewDecimal("123")
	assert.NoError(t, err)
	data := invoice.Data{
		Sender:      did,
		Number:      "2132131",
		Status:      "unpaid",
		GrossAmount: dec,
		NetAmount:   dec,
		Currency:    "EUR",
		DateDue:     tm,
	}

	d, err := json.Marshal(data)
	assert.NoError(t, err)
	return d
}
