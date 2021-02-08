// +build integration

package nft_test

import (
	"context"
	"os"
	"sync"
	"testing"

	"github.com/centrifuge/centrifuge-protobufs/documenttypes"
	"github.com/centrifuge/go-centrifuge/bootstrap"
	cc "github.com/centrifuge/go-centrifuge/bootstrap/bootstrappers/testingbootstrap"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/config/configstore"
	"github.com/centrifuge/go-centrifuge/contextutil"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/documents/generic"
	"github.com/centrifuge/go-centrifuge/ethereum"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/jobs"
	"github.com/centrifuge/go-centrifuge/jobs/jobsv2"
	"github.com/centrifuge/go-centrifuge/nft"
	"github.com/centrifuge/go-centrifuge/testingutils"
	testingidentity "github.com/centrifuge/go-centrifuge/testingutils/identity"
	"github.com/centrifuge/gocelery/v2"
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
var nftService nft.Service
var jobManager jobs.Manager
var tokenRegistry documents.TokenRegistry
var dispatcher jobsv2.Dispatcher
var ethClient ethereum.Client

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
	dispatcher = ctx[jobsv2.BootstrappedDispatcher].(jobsv2.Dispatcher)
	ethClient = ctx[ethereum.BootstrappedEthereumClient].(ethereum.Client)
	ctxh, canc := context.WithCancel(context.Background())
	wg := new(sync.WaitGroup)
	wg.Add(1)
	go dispatcher.Start(ctxh, wg, nil)
	result := m.Run()
	cc.TestFunctionalEthereumTearDown()
	canc()
	wg.Wait()
	os.Exit(result)
}

func createIdentity(t *testing.T) (identity.DID, config.Account) {
	// create identity
	log.Debug("Create Identity for Testing")
	did, err := idFactory.NextIdentityAddress()
	assert.NoError(t, err)
	tc, err := configstore.TempAccount("main", cfg)
	assert.NoError(t, err)
	tcr := tc.(*configstore.Account)
	tcr.IdentityID = did[:]
	_, err = cfgService.CreateAccount(tcr)
	assert.NoError(t, err)
	cid, err := testingidentity.CreateAccountIDWithKeys(
		cfg.GetEthereumContextWaitTimeout(), tcr, idService, idFactory, ethClient, dispatcher)
	assert.NoError(t, err)
	return cid, tcr
}

func prepareGenericForNFTMinting(
	t *testing.T,
	regAddr string,
	cid identity.DID,
	tcr config.Account,
	attrs map[documents.AttrKey]documents.Attribute) (context.Context, []byte, common.Address, documents.Service, identity.DID) {
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

func mintNFT(t *testing.T, ctx context.Context, req nft.MintNFTRequest, did identity.DID, registry common.Address) nft.TokenID {
	resp, err := nftService.MintNFT(ctx, req)
	assert.NoError(t, err, "should not error out when minting an invoice")
	assert.NotNil(t, resp.TokenID, "token id should be present")
	tokenID, err := nft.TokenIDFromString(resp.TokenID)
	assert.NoError(t, err, "should not error out when getting tokenID hex")
	jobID := hexutil.MustDecode(resp.JobID)
	res, err := dispatcher.Result(did, jobID)
	assert.NoError(t, err)
	_, err = res.Await(context.Background())
	assert.NoError(t, err)
	owner, err := tokenRegistry.OwnerOf(registry, tokenID.BigInt().Bytes())
	assert.NoError(t, err)
	assert.Equal(t, req.DepositAddress, owner)
	return tokenID
}

func mintNFTWithProofs(t *testing.T) (context.Context, nft.TokenID, identity.DID) {
	did, acc := createIdentity(t)
	attrs, pfs := nft.GetAttributes(t, did)
	scAddrs := testingutils.GetDAppSmartContractAddresses()
	ctx, id, registry, invSrv, cid := prepareGenericForNFTMinting(t, scAddrs["genericNFT"], did, acc, attrs)
	pfs = append(pfs, nft.GetSignatureProofField(t, acc))
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
	resp, err := nftService.TransferFrom(ctx, registry, to, tokenID)
	assert.NoError(t, err)
	jobID := gocelery.JobID(hexutil.MustDecode(resp.JobID))
	res, err := dispatcher.Result(did, jobID)
	assert.NoError(t, err)
	_, err = res.Await(ctx)
	assert.NoError(t, err)

	owner, err = nftService.OwnerOf(registry, tokenID[:])
	assert.NoError(t, err)
	assert.Equal(t, owner, to)

	// should fail not owner anymore
	secondTo := common.HexToAddress("0xFBb1b73C4f0BDa4f67dcA266ce6Ef42f520fBB98")
	resp, err = nftService.TransferFrom(ctx, registry, secondTo, tokenID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "is not the owner of NFT")
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
