//go:build integration

package v3_test

import (
	"context"
	"os"
	"sync"
	"testing"

	"github.com/centrifuge/centrifuge-protobufs/documenttypes"
	coredocumentpb "github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/config/configstore"
	"github.com/centrifuge/go-centrifuge/contextutil"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/documents/generic"
	v2 "github.com/centrifuge/go-centrifuge/identity/v2"
	"github.com/centrifuge/go-centrifuge/jobs"
	nftv3 "github.com/centrifuge/go-centrifuge/nft/v3"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/centrifuge/gocelery/v2"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/stretchr/testify/assert"
)

var (
	nftV3Srv   nftv3.Service
	registry   *documents.ServiceRegistry
	cfg        config.Configuration
	cfgService config.Service
	idService  v2.Service
	dispatcher jobs.Dispatcher
)

func TestMain(m *testing.M) {
	testCtx := testingbootstrap.TestFunctionalEthereumBootstrap()
	nftV3Srv = testCtx[bootstrap.BootstrappedNFTV3Service].(nftv3.Service)
	registry = testCtx[documents.BootstrappedRegistry].(*documents.ServiceRegistry)
	idService = testCtx[identity.BootstrappedDIDService].(identity.Service)
	idFactory = testCtx[identity.BootstrappedDIDFactory].(identity.Factory)
	cfg = testCtx[bootstrap.BootstrappedConfig].(config.Configuration)
	cfgService = testCtx[config.BootstrappedConfigStorage].(config.Service)
	dispatcher = testCtx[jobs.BootstrappedDispatcher].(jobs.Dispatcher)
	ethClient = testCtx[ethereum.BootstrappedEthereumClient].(ethereum.Client)

	ctx, canc := context.WithCancel(context.Background())

	var wg sync.WaitGroup
	wg.Add(1)

	go dispatcher.Start(ctx, &wg, nil)
	result := m.Run()
	testingbootstrap.TestFunctionalEthereumTearDown()
	canc()
	wg.Wait()
	os.Exit(result)
}

func TestService_Mint(t *testing.T) {
	did, acc := createIdentity(t)
	ctx := contextutil.WithAccount(context.Background(), acc)

	docID, err := createGenericDocument(t, ctx, did)
	assert.NoError(t, err)

	classID := types.U64(1234)

	createClassReq := &nftv3.CreateNFTClassRequest{
		ClassID: classID,
	}

	createClassRes, err := nftV3Srv.CreateNFTClass(ctx, createClassReq)
	assert.NoError(t, err)
	assert.NotNil(t, createClassRes)

	jobID := gocelery.JobID(hexutil.MustDecode(createClassRes.JobID))
	result, err := dispatcher.Result(did, jobID)
	assert.NoError(t, err)

	_, err = result.Await(ctx)
	assert.NoError(t, err)

	nftMeta := "metadata_test"

	mintReq := &nftv3.MintNFTRequest{
		DocumentID:     docID,
		Metadata:       nftMeta,
		FreezeMetadata: true,
		ClassID:        classID,
		Owner:          types.NewAccountID([]byte(acc.GetCentChainAccount().ID)),
	}

	mintRes, err := nftV3Srv.MintNFT(ctx, mintReq)
	assert.NoError(t, err)
	assert.NotNil(t, mintRes)

	jobID = hexutil.MustDecode(mintRes.JobID)
	result, err = dispatcher.Result(did, jobID)
	assert.NoError(t, err)

	_, err = result.Await(ctx)
	assert.NoError(t, err)

	doc, err := getGenericDocument(t, ctx, docID)
	assert.NoError(t, err)

	var nft *coredocumentpb.CcNft

	for _, ccNft := range doc.CcNfts() {
		var nftClassID types.U64

		if err := types.DecodeFromBytes(ccNft.GetClassId(), &nftClassID); err != nil {
			t.Fatalf("Couldn't decode class ID from CC NFT: %s", err)
		}

		if nftClassID == classID {
			nft = ccNft
			break
		}
	}

	assert.NotNil(t, nft)

	var instanceID types.U128

	err = types.DecodeFromBytes(nft.GetInstanceId(), &instanceID)
	assert.NoError(t, err)

	ownerReq := &nftv3.OwnerOfRequest{
		ClassID:    classID,
		InstanceID: instanceID,
	}

	ownerRes, err := nftV3Srv.OwnerOf(ctx, ownerReq)
	assert.NoError(t, err)
	assert.Equal(t, types.NewAccountID([]byte(acc.GetCentChainAccount().ID)), ownerRes.AccountID)

	instanceMetadataReq := &nftv3.ItemMetadataOfRequest{
		ClassID:    classID,
		InstanceID: instanceID,
	}

	instanceMetaRes, err := nftV3Srv.InstanceMetadataOf(ctx, instanceMetadataReq)
	assert.NoError(t, err)
	assert.Equal(t, nftMeta, string(instanceMetaRes.Data))
	assert.True(t, instanceMetaRes.IsFrozen)
}

func TestService_Mint_MissingClass(t *testing.T) {
	did, acc := createIdentity(t)
	ctx := contextutil.WithAccount(context.Background(), acc)

	docID, err := createGenericDocument(t, ctx, did)
	assert.NoError(t, err)

	classID := types.U64(1234)

	nftMeta := "metadata_test"

	mintReq := &nftv3.MintNFTRequest{
		DocumentID:     docID,
		Metadata:       nftMeta,
		FreezeMetadata: true,
		ClassID:        classID,
		Owner:          types.NewAccountID([]byte(acc.GetCentChainAccount().ID)),
	}

	mintRes, err := nftV3Srv.MintNFT(ctx, mintReq)
	assert.NoError(t, err)
	assert.NotNil(t, mintRes)

	jobID := hexutil.MustDecode(mintRes.JobID)
	result, err := dispatcher.Result(did, jobID)
	assert.NoError(t, err)

	_, err = result.Await(ctx)
	assert.NotNil(t, err)
}

func TestService_CreateNFTClass(t *testing.T) {
	did, acc := createIdentity(t)
	ctx := contextutil.WithAccount(context.Background(), acc)

	classID := types.U64(1234)

	createClassReq := &nftv3.CreateNFTClassRequest{
		ClassID: classID,
	}

	createClassRes, err := nftV3Srv.CreateNFTClass(ctx, createClassReq)
	assert.NoError(t, err)
	assert.NotNil(t, createClassRes)

	jobID := gocelery.JobID(hexutil.MustDecode(createClassRes.JobID))
	result, err := dispatcher.Result(did, jobID)
	assert.NoError(t, err)

	_, err = result.Await(ctx)
	assert.NoError(t, err)

	// Class already exists
	createClassRes, err = nftV3Srv.CreateNFTClass(ctx, createClassReq)
	assert.NotNil(t, err)
	assert.Nil(t, createClassRes)
}

func createGenericDocument(t *testing.T, ctx context.Context, did identity.DID) ([]byte, error) {
	genericSrv, err := registry.LocateService(documenttypes.GenericDataTypeUrl)
	assert.NoError(t, err)

	doc, err := genericSrv.Derive(
		ctx,
		documents.UpdatePayload{
			CreatePayload: documents.CreatePayload{
				Scheme: generic.Scheme,
				Collaborators: documents.CollaboratorsAccess{
					ReadWriteCollaborators: nil,
				},
				Attributes: nil,
			},
		})
	assert.NoError(t, err)

	jobID, err := genericSrv.Commit(ctx, doc)
	assert.NoError(t, err)

	res, err := dispatcher.Result(did, jobID)
	assert.NoError(t, err)

	_, err = res.Await(ctx)
	assert.NoError(t, err)

	return doc.ID(), err
}

func getGenericDocument(t *testing.T, ctx context.Context, docID []byte) (documents.Document, error) {
	genericSrv, err := registry.LocateService(documenttypes.GenericDataTypeUrl)
	assert.NoError(t, err)

	return genericSrv.GetCurrentVersion(ctx, docID)
}

func createIdentity(t *testing.T) (identity.DID, config.Account) {
	did, err := idFactory.NextIdentityAddress()
	assert.NoError(t, err)

	tc, err := configstore.TempAccount("main", cfg)
	assert.NoError(t, err)

	tcr := tc.(*configstore.Account)
	tcr.IdentityID = did[:]

	_, err = cfgService.CreateAccount(tcr)
	assert.NoError(t, err)

	cid, err := testingidentity.CreateAccountIDWithKeys(
		cfg.GetEthereumContextWaitTimeout(),
		tcr,
		idService,
		idFactory,
		ethClient,
		dispatcher,
	)
	assert.NoError(t, err)

	return cid, tcr
}
