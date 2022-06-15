//go:build integration
// +build integration

package nft_test

import (
	"context"
	"fmt"
	"os"
	"sync"
	"testing"

	"github.com/centrifuge/centrifuge-protobufs/documenttypes"
	"github.com/centrifuge/go-centrifuge/bootstrap"
	cc "github.com/centrifuge/go-centrifuge/bootstrap/bootstrappers/testingbootstrap"
	"github.com/centrifuge/go-centrifuge/centchain"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/config/configstore"
	"github.com/centrifuge/go-centrifuge/contextutil"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/documents/generic"
	"github.com/centrifuge/go-centrifuge/ethereum"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/jobs"
	"github.com/centrifuge/go-centrifuge/nft"
	"github.com/centrifuge/go-centrifuge/testingutils"
	testingidentity "github.com/centrifuge/go-centrifuge/testingutils/identity"
	"github.com/centrifuge/go-substrate-rpc-client/v4/signature"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
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
var tokenRegistry documents.TokenRegistry
var dispatcher jobs.Dispatcher
var ethClient ethereum.Client
var centAPI centchain.API

func TestMain(m *testing.M) {
	log.Debug("Test PreSetup for NFT")
	ctx := cc.TestFunctionalEthereumBootstrap()
	registry = ctx[documents.BootstrappedRegistry].(*documents.ServiceRegistry)
	idService = ctx[identity.BootstrappedDIDService].(identity.Service)
	idFactory = ctx[identity.BootstrappedDIDFactory].(identity.Factory)
	cfg = ctx[bootstrap.BootstrappedConfig].(config.Configuration)
	cfgService = ctx[config.BootstrappedConfigStorage].(config.Service)
	nftService = ctx[bootstrap.BootstrappedNFTService].(nft.Service)
	tokenRegistry = ctx[bootstrap.BootstrappedNFTService].(documents.TokenRegistry)
	dispatcher = ctx[jobs.BootstrappedDispatcher].(jobs.Dispatcher)
	ethClient = ctx[ethereum.BootstrappedEthereumClient].(ethereum.Client)
	centAPI = ctx[centchain.BootstrappedCentChainClient].(centchain.API)
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
	cid identity.DID,
	tcr config.Account,
	attrs map[documents.AttrKey]documents.Attribute) (context.Context, []byte, documents.Service) {
	// create Generic doc (anchor)
	genericSrv, err := registry.LocateService(documenttypes.GenericDataTypeUrl)
	assert.Nil(t, err, "should not error out when getting generic genSrv")
	ctx, err := contextutil.New(context.Background(), tcr)
	assert.NoError(t, err)

	payload := genericPayload(t, nil, attrs)
	doc, err := genericSrv.Derive(ctx, documents.UpdatePayload{CreatePayload: payload})
	assert.NoError(t, err)
	jobID, err := genericSrv.Commit(ctx, doc)
	assert.NoError(t, err)
	res, err := dispatcher.Result(cid, jobID)
	assert.NoError(t, err)
	_, err = res.Await(ctx)
	assert.NoError(t, err)
	return ctx, doc.ID(), genericSrv
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
	registry := common.HexToAddress(scAddrs["genericNFT"])
	ctx, id, docSrv := prepareGenericForNFTMinting(t, did, acc, attrs)
	pfs = append(pfs, nft.GetSignatureProofField(t, acc))
	req := nft.MintNFTRequest{
		DocumentID:               id,
		RegistryAddress:          registry,
		DepositAddress:           did.ToAddress(),
		AssetManagerAddress:      common.HexToAddress(scAddrs["assetManager"]),
		ProofFields:              pfs,
		GrantNFTReadAccess:       false,
		SubmitNFTReadAccessProof: false,
		SubmitTokenProof:         true,
	}
	tokenID := mintNFT(t, ctx, req, did, registry)
	_, err := docSrv.GetCurrentVersion(ctx, id)
	assert.NoError(t, err)
	return ctx, tokenID, did
}

func TestTransferNFT(t *testing.T) {
	t.SkipNow() // TODO Re-enable once we have the new NFT integration in
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

func TestMintCCNFTAndTransfer(t *testing.T) {
	t.SkipNow() // TODO Re-enable once we have the new NFT integration in
	did, acc := createIdentity(t)
	ctx := contextutil.WithAccount(context.Background(), acc)

	proofFields := [][]byte{
		// originator
		hexutil.MustDecode("0x010000000000001ce24e7917d4fcaf79095539ac23af9f6d5c80ea8b0d95c9cd860152bff8fdab1700000005"),
		// asset value
		hexutil.MustDecode("0x010000000000001ccd35852d8705a28d4f83ba46f02ebdf46daf03638b40da74b9371d715976e6dd00000005"),
		// asset identifier
		hexutil.MustDecode("0x010000000000001cbbaa573c53fa357a3b53624eb6deab5f4c758f299cffc2b0b6162400e3ec13ee00000005"),
		// MaturityDate
		hexutil.MustDecode("0x010000000000001ce5588a8a267ed4c32962568afe216d4ba70ae60576a611e3ca557b84f1724e2900000005"),
	}
	info := nft.RegistryInfo{
		OwnerCanBurn: true,
		Fields:       proofFields,
	}

	api := nft.NewAPI(centAPI)
	registry, err := api.CreateRegistry(ctx, info)
	assert.NoError(t, err)
	assert.NotEmpty(t, registry)
	fmt.Println("NFT registry:", registry.Hex())

	owner := types.NewAccountID(signature.TestKeyringPairAlice.PublicKey)
	attrs, pfs := nft.GetAttributes(t, did)
	ctx, docID, _ := prepareGenericForNFTMinting(t, did, acc, attrs)

	req := nft.MintNFTOnCCRequest{
		DocumentID:         docID,
		ProofFields:        pfs,
		RegistryAddress:    registry,
		DepositAddress:     owner,
		GrantNFTReadAccess: true,
	}

	resp, err := nftService.MintNFTOnCC(ctx, req)
	assert.NoError(t, err)

	jobID := gocelery.JobID(hexutil.MustDecode(resp.JobID))
	result, err := dispatcher.Result(did, jobID)
	assert.NoError(t, err)
	_, err = result.Await(ctx)
	assert.NoError(t, err)

	tokenID, err := nft.TokenIDFromString(resp.TokenID)
	assert.NoError(t, err)

	fmt.Printf("NFT minted: Registry[%s]->Token[%s]\n", registry.Hex(), tokenID.String())
	cown, err := nftService.OwnerOfOnCC(registry, tokenID[:])
	assert.NoError(t, err)
	assert.Equal(t, owner, cown)
	fmt.Printf("NFT owner verified: Token[%s]->Owner[%s]\n", tokenID.String(), hexutil.Encode(cown[:]))

	kr, err := signature.KeyringPairFromSecret("//Bob", 42)
	assert.NoError(t, err)
	resp, err = nftService.TransferNFT(ctx, registry, tokenID, types.NewAccountID(kr.PublicKey))
	assert.NoError(t, err)
	jobID = hexutil.MustDecode(resp.JobID)
	result, err = dispatcher.Result(did, jobID)
	assert.NoError(t, err)
	_, err = result.Await(ctx)
	assert.NoError(t, err)
	fmt.Printf("NFT transferred Token[%s]: Old Owner[%s]->New Owner[%s]\n", tokenID.String(), hexutil.Encode(cown[:]),
		hexutil.Encode(kr.PublicKey))
}
