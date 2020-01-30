// +build integration

package nft_test

import (
	"context"
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
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/jobs"
	"github.com/centrifuge/go-centrifuge/nft"
	"github.com/centrifuge/go-centrifuge/testingutils"
	"github.com/centrifuge/go-centrifuge/testingutils/identity"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
	"github.com/stretchr/testify/assert"
)

var registry *documents.ServiceRegistry
var cfg config.Configuration
var cfgService config.Service
var idService identity.Service
var idFactory identity.Factory
var nftService nft.Service
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
	nftService = ctx[bootstrap.BootstrappedNFTService].(nft.Service)
	jobManager = ctx[jobs.BootstrappedService].(jobs.Manager)
	tokenRegistry = ctx[bootstrap.BootstrappedNFTService].(documents.TokenRegistry)
	result := m.Run()
	cc.TestFunctionalEthereumTearDown()
	os.Exit(result)
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
	resp, done, err := nftService.MintNFT(ctx, req)
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

func getAttributes(t *testing.T) (map[documents.AttrKey]documents.Attribute, []string) {
	attrs := map[documents.AttrKey]documents.Attribute{}
	loanAmount := "loanAmount"
	loanAmountValue := "100.10001"
	attr0, err := documents.NewStringAttribute(loanAmount, documents.AttrDecimal, loanAmountValue)
	assert.NoError(t, err)
	attrs[attr0.Key] = attr0
	asIsValue := "dateValue"
	asIsValueValue := time.Now().UTC().Format(time.RFC3339Nano)
	attr1, err := documents.NewStringAttribute(asIsValue, documents.AttrTimestamp, asIsValueValue)
	assert.NoError(t, err)
	attrs[attr1.Key] = attr1
	afterRehabValue := "afterRehabValue"
	afterRehabValueValue := "2000"
	attr2, err := documents.NewStringAttribute(afterRehabValue, documents.AttrDecimal, afterRehabValueValue)
	assert.NoError(t, err)
	attrs[attr2.Key] = attr2

	attributeLoanAmount := fmt.Sprintf("%s.attributes[%s].byte_val", documents.CDTreePrefix, attr0.Key.String())
	attributeAsIsVal := fmt.Sprintf("%s.attributes[%s].byte_val", documents.CDTreePrefix, attr1.Key.String())
	attributeAfterRehabVal := fmt.Sprintf("%s.attributes[%s].byte_val", documents.CDTreePrefix, attr2.Key.String())
	proofFields := []string{attributeLoanAmount, attributeAsIsVal, attributeAfterRehabVal}
	return attrs, proofFields
}

func mintNFTWithProofs(t *testing.T) (context.Context, nft.TokenID, identity.DID) {
	attrs, pfs := getAttributes(t)
	scAddrs := testingutils.GetDAppSmartContractAddresses()
	ctx, id, registry, invSrv, cid := prepareGenericForNFTMinting(t, scAddrs["genericNFT"], attrs)
	req := nft.MintNFTRequest{
		DocumentID:               id,
		RegistryAddress:          registry,
		DepositAddress:           cid.ToAddress(),
		AssetManagerAddress:      common.HexToAddress(scAddrs["assetManager"]),
		ProofFields:              pfs,
		GrantNFTReadAccess:       false,
		SubmitNFTReadAccessProof: false,
		SubmitTokenProof:         true,
	}
	tokenID := mintNFT(t, ctx, req, cid, registry)
	doc, err := invSrv.GetCurrentVersion(ctx, id)
	assert.NoError(t, err)
	cd, err := doc.PackCoreDocument()
	assert.NoError(t, err)
	assert.Len(t, cd.Roles, 2)
	return ctx, tokenID, cid
}

func TestTransferNFT(t *testing.T) {
	scAddrs := testingutils.GetDAppSmartContractAddresses()
	registry := common.HexToAddress(scAddrs["genericNFT"])
	ctx, tokenID, did := mintNFTWithProofs(t)
	to := common.HexToAddress("0xde0B295669a9FD93d5F28D9Ec85E40f4cb697BAe")
	owner, err := nftService.OwnerOf(registry, tokenID[:])
	assert.NoError(t, err)
	assert.Equal(t, owner, did.ToAddress())

	// successful
	resp, done, err := nftService.TransferFrom(ctx, registry, to, tokenID)
	assert.NoError(t, err)
	err = <-done
	assert.NoError(t, err)
	jobID, err := jobs.FromString(resp.JobID)
	assert.NoError(t, err)
	assert.NoError(t, jobManager.WaitForJob(did, jobID))

	owner, err = nftService.OwnerOf(registry, tokenID[:])
	assert.NoError(t, err)
	assert.Equal(t, owner, to)

	// should fail not owner anymore
	secondTo := common.HexToAddress("0xFBb1b73C4f0BDa4f67dcA266ce6Ef42f520fBB98")
	resp, done, err = nftService.TransferFrom(ctx, registry, secondTo, tokenID)
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

func genericPayload(t *testing.T, collaborators []identity.DID, attrs map[documents.AttrKey]documents.Attribute) documents.CreatePayload {
	return documents.CreatePayload{
		Scheme: generic.Scheme,
		Collaborators: documents.CollaboratorsAccess{
			ReadWriteCollaborators: collaborators,
		},
		Attributes: attrs,
	}
}
