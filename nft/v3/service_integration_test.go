//go:build integration

package v3_test

import (
	"context"
	"encoding/json"
	"math/rand"
	"os"
	"testing"

	"github.com/centrifuge/go-centrifuge/utils"

	"github.com/ipfs/go-cid"
	"github.com/ipfs/interface-go-ipfs-core/path"
	mh "github.com/multiformats/go-multihash"

	"github.com/centrifuge/centrifuge-protobufs/documenttypes"
	coredocumentpb "github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/go-centrifuge/anchors"
	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/bootstrap/bootstrappers/integration_test"
	"github.com/centrifuge/go-centrifuge/bootstrap/bootstrappers/testlogging"
	"github.com/centrifuge/go-centrifuge/centchain"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/config/configstore"
	"github.com/centrifuge/go-centrifuge/contextutil"
	protocolIDDispatcher "github.com/centrifuge/go-centrifuge/dispatcher"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/documents/generic"
	v2 "github.com/centrifuge/go-centrifuge/identity/v2"
	"github.com/centrifuge/go-centrifuge/ipfs_pinning"
	"github.com/centrifuge/go-centrifuge/jobs"
	nftv3 "github.com/centrifuge/go-centrifuge/nft/v3"
	"github.com/centrifuge/go-centrifuge/p2p"
	"github.com/centrifuge/go-centrifuge/pallets"
	"github.com/centrifuge/go-centrifuge/pending"
	"github.com/centrifuge/go-centrifuge/storage/leveldb"
	"github.com/centrifuge/go-centrifuge/testingutils/keyrings"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types/codec"
	"github.com/centrifuge/gocelery/v2"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/stretchr/testify/assert"
)

var integrationTestBootstrappers = []bootstrap.TestBootstrapper{
	&testlogging.TestLoggingBootstrapper{},
	&config.Bootstrapper{},
	&leveldb.Bootstrapper{},
	jobs.Bootstrapper{},
	&configstore.Bootstrapper{},
	&integration_test.Bootstrapper{},
	centchain.Bootstrapper{},
	&pallets.Bootstrapper{},
	&protocolIDDispatcher.Bootstrapper{},
	&v2.Bootstrapper{},
	anchors.Bootstrapper{},
	documents.Bootstrapper{},
	pending.Bootstrapper{},
	&ipfs_pinning.Bootstrapper{},
	&nftv3.Bootstrapper{},
	p2p.Bootstrapper{},
	documents.PostBootstrapper{},
	generic.Bootstrapper{},
}

var (
	nftService    nftv3.Service
	registry      *documents.ServiceRegistry
	dispatcher    jobs.Dispatcher
	cfgService    config.Service
	pendingDocSrv pending.Service
)

func TestMain(m *testing.M) {
	ctx := bootstrap.RunTestBootstrappers(integrationTestBootstrappers, nil)
	nftService = ctx[nftv3.BootstrappedNFTV3Service].(nftv3.Service)
	registry = ctx[documents.BootstrappedRegistry].(*documents.ServiceRegistry)
	dispatcher = ctx[jobs.BootstrappedJobDispatcher].(jobs.Dispatcher)
	cfgService = ctx[config.BootstrappedConfigStorage].(config.Service)
	pendingDocSrv = ctx[pending.BootstrappedPendingDocumentService].(pending.Service)

	result := m.Run()

	bootstrap.RunTestTeardown(integrationTestBootstrappers)

	os.Exit(result)
}

func TestIntegration_Service_MintNFT_NonPendingDocument(t *testing.T) {
	acc, err := cfgService.GetAccount(keyrings.AliceKeyRingPair.PublicKey)
	assert.NoError(t, err)

	ctx := contextutil.WithAccount(context.Background(), acc)

	docID, err := createAndCommitGenericDocument(t, ctx, acc.GetIdentity())
	assert.NoError(t, err)

	collectionID := types.U64(rand.Int63())

	createCollectionReq := &nftv3.CreateNFTCollectionRequest{
		CollectionID: collectionID,
	}

	createCollectionRes, err := nftService.CreateNFTCollection(ctx, createCollectionReq)
	assert.NoError(t, err)
	assert.NotNil(t, createCollectionRes)

	jobID := gocelery.JobID(hexutil.MustDecode(createCollectionRes.JobID))
	result, err := dispatcher.Result(acc.GetIdentity(), jobID)
	assert.NoError(t, err)

	_, err = result.Await(ctx)
	assert.NoError(t, err)

	docAttrKeyLabels := []string{
		"test-label-1",
		"test-label-2",
	}

	ipfsMeta := nftv3.IPFSMetadata{
		Name:                  "test-name",
		Description:           "test-desc",
		Image:                 "test-image",
		DocumentAttributeKeys: docAttrKeyLabels,
	}

	mintReq := &nftv3.MintNFTRequest{
		DocumentID:      docID,
		CollectionID:    collectionID,
		Owner:           acc.GetIdentity(),
		IPFSMetadata:    ipfsMeta,
		GrantReadAccess: false,
	}

	mintRes, err := nftService.MintNFT(ctx, mintReq, false)
	assert.NoError(t, err)
	assert.NotNil(t, mintRes)

	jobID = hexutil.MustDecode(mintRes.JobID)
	result, err = dispatcher.Result(acc.GetIdentity(), jobID)
	assert.NoError(t, err)

	_, err = result.Await(ctx)
	assert.NoError(t, err)

	doc, err := getGenericDocument(t, ctx, docID)
	assert.NoError(t, err)

	var nft *coredocumentpb.NFT

	for _, docNft := range doc.NFTs() {
		var nftCollectionID types.U64

		if err := codec.Decode(docNft.GetCollectionId(), &nftCollectionID); err != nil {
			t.Fatalf("Couldn't decode class ID from CC NFT: %s", err)
		}

		if nftCollectionID == collectionID {
			nft = docNft
			break
		}
	}

	assert.NotNil(t, nft)

	var itemID types.U128

	err = codec.Decode(nft.GetItemId(), &itemID)
	assert.NoError(t, err)

	ownerReq := &nftv3.GetNFTOwnerRequest{
		CollectionID: collectionID,
		ItemID:       itemID,
	}

	ownerRes, err := nftService.GetNFTOwner(ctx, ownerReq)
	assert.NoError(t, err)
	assert.Equal(t, acc.GetIdentity(), ownerRes.AccountID)

	instanceMetadataReq := &nftv3.GetItemMetadataRequest{
		CollectionID: collectionID,
		ItemID:       itemID,
	}

	instanceMetaRes, err := nftService.GetItemMetadata(ctx, instanceMetadataReq)
	assert.NoError(t, err)
	assert.False(t, instanceMetaRes.IsFrozen)

	docAttrMap, err := nftv3.GetDocAttributes(doc, docAttrKeyLabels)
	assert.NoError(t, err)

	nftMeta := ipfs_pinning.NFTMetadata{
		Name:        ipfsMeta.Name,
		Description: ipfsMeta.Description,
		Image:       ipfsMeta.Image,
		Properties:  docAttrMap,
	}

	nftMetaJSON, err := json.Marshal(nftMeta)
	assert.NoError(t, err)

	v1CidPrefix := cid.Prefix{
		Codec:    cid.Raw,
		MhLength: -1,
		MhType:   mh.SHA2_256,
		Version:  1,
	}

	metadataCID, err := v1CidPrefix.Sum(nftMetaJSON)
	assert.NoError(t, err)

	metaPath := path.New(metadataCID.String())

	assert.Equal(t, metaPath.String(), string(instanceMetaRes.Data))
}

func TestIntegration_Service_MintNFT_PendingDocument(t *testing.T) {
	acc, err := cfgService.GetAccount(keyrings.AliceKeyRingPair.PublicKey)
	assert.NoError(t, err)

	ctx := contextutil.WithAccount(context.Background(), acc)

	doc, err := createPendingDocument(t, ctx)
	assert.NoError(t, err)

	collectionID := types.U64(rand.Int63())

	createClassReq := &nftv3.CreateNFTCollectionRequest{
		CollectionID: collectionID,
	}

	createClassRes, err := nftService.CreateNFTCollection(ctx, createClassReq)
	assert.NoError(t, err)
	assert.NotNil(t, createClassRes)

	jobID := gocelery.JobID(hexutil.MustDecode(createClassRes.JobID))
	result, err := dispatcher.Result(acc.GetIdentity(), jobID)
	assert.NoError(t, err)

	_, err = result.Await(ctx)
	assert.NoError(t, err)

	docAttrKeyLabels := []string{
		"test-label-1",
		"test-label-2",
	}

	ipfsMeta := nftv3.IPFSMetadata{
		Name:                  "test-name",
		Description:           "test-desc",
		Image:                 "test-image",
		DocumentAttributeKeys: docAttrKeyLabels,
	}

	mintReq := &nftv3.MintNFTRequest{
		DocumentID:      doc.ID(),
		CollectionID:    collectionID,
		Owner:           acc.GetIdentity(),
		IPFSMetadata:    ipfsMeta,
		GrantReadAccess: false,
	}

	mintRes, err := nftService.MintNFT(ctx, mintReq, true)
	assert.NoError(t, err)
	assert.NotNil(t, mintRes)

	jobID = hexutil.MustDecode(mintRes.JobID)
	result, err = dispatcher.Result(acc.GetIdentity(), jobID)
	assert.NoError(t, err)

	_, err = result.Await(ctx)
	assert.NoError(t, err)

	doc, err = getGenericDocument(t, ctx, doc.ID())
	assert.NoError(t, err)

	var nft *coredocumentpb.NFT

	for _, docNft := range doc.NFTs() {
		var nftCollectionID types.U64

		if err := codec.Decode(docNft.GetCollectionId(), &nftCollectionID); err != nil {
			t.Fatalf("Couldn't decode class ID from CC NFT: %s", err)
		}

		if nftCollectionID == collectionID {
			nft = docNft
			break
		}
	}

	assert.NotNil(t, nft)

	var itemID types.U128

	err = codec.Decode(nft.GetItemId(), &itemID)
	assert.NoError(t, err)

	ownerReq := &nftv3.GetNFTOwnerRequest{
		CollectionID: collectionID,
		ItemID:       itemID,
	}

	ownerRes, err := nftService.GetNFTOwner(ctx, ownerReq)
	assert.NoError(t, err)
	assert.Equal(t, acc.GetIdentity(), ownerRes.AccountID)

	instanceMetadataReq := &nftv3.GetItemMetadataRequest{
		CollectionID: collectionID,
		ItemID:       itemID,
	}

	instanceMetaRes, err := nftService.GetItemMetadata(ctx, instanceMetadataReq)
	assert.NoError(t, err)
	assert.False(t, instanceMetaRes.IsFrozen)

	docAttrMap, err := nftv3.GetDocAttributes(doc, docAttrKeyLabels)
	assert.NoError(t, err)

	nftMeta := ipfs_pinning.NFTMetadata{
		Name:        ipfsMeta.Name,
		Description: ipfsMeta.Description,
		Image:       ipfsMeta.Image,
		Properties:  docAttrMap,
	}

	nftMetaJSON, err := json.Marshal(nftMeta)
	assert.NoError(t, err)

	v1CidPrefix := cid.Prefix{
		Codec:    cid.Raw,
		MhLength: -1,
		MhType:   mh.SHA2_256,
		Version:  1,
	}

	metadataCID, err := v1CidPrefix.Sum(nftMetaJSON)
	assert.NoError(t, err)

	metaPath := path.New(metadataCID.String())

	assert.Equal(t, metaPath.String(), string(instanceMetaRes.Data))
}

func TestIntegration_Service_MintNFT_NonPendingDocument_DocumentNotPresent(t *testing.T) {
	acc, err := cfgService.GetAccount(keyrings.AliceKeyRingPair.PublicKey)
	assert.NoError(t, err)

	ctx := contextutil.WithAccount(context.Background(), acc)

	docID := utils.RandomSlice(32)

	collectionID := types.U64(rand.Int63())

	createCollectionReq := &nftv3.CreateNFTCollectionRequest{
		CollectionID: collectionID,
	}

	createCollectionRes, err := nftService.CreateNFTCollection(ctx, createCollectionReq)
	assert.NoError(t, err)
	assert.NotNil(t, createCollectionRes)

	jobID := gocelery.JobID(hexutil.MustDecode(createCollectionRes.JobID))
	result, err := dispatcher.Result(acc.GetIdentity(), jobID)
	assert.NoError(t, err)

	_, err = result.Await(ctx)
	assert.NoError(t, err)

	docAttrKeyLabels := []string{
		"test-label-1",
		"test-label-2",
	}

	ipfsMeta := nftv3.IPFSMetadata{
		Name:                  "test-name",
		Description:           "test-desc",
		Image:                 "test-image",
		DocumentAttributeKeys: docAttrKeyLabels,
	}

	mintReq := &nftv3.MintNFTRequest{
		DocumentID:      docID,
		CollectionID:    collectionID,
		Owner:           acc.GetIdentity(),
		IPFSMetadata:    ipfsMeta,
		GrantReadAccess: false,
	}

	mintRes, err := nftService.MintNFT(ctx, mintReq, false)
	assert.ErrorIs(t, err, nftv3.ErrDocumentRetrieval)
	assert.Nil(t, mintRes)
}

func TestIntegration_Service_MintNFT_PendingDocument_DocumentNotPresent(t *testing.T) {
	acc, err := cfgService.GetAccount(keyrings.AliceKeyRingPair.PublicKey)
	assert.NoError(t, err)

	ctx := contextutil.WithAccount(context.Background(), acc)

	docID := utils.RandomSlice(32)

	collectionID := types.U64(rand.Int63())

	createClassReq := &nftv3.CreateNFTCollectionRequest{
		CollectionID: collectionID,
	}

	createClassRes, err := nftService.CreateNFTCollection(ctx, createClassReq)
	assert.NoError(t, err)
	assert.NotNil(t, createClassRes)

	jobID := gocelery.JobID(hexutil.MustDecode(createClassRes.JobID))
	result, err := dispatcher.Result(acc.GetIdentity(), jobID)
	assert.NoError(t, err)

	_, err = result.Await(ctx)
	assert.NoError(t, err)

	docAttrKeyLabels := []string{
		"test-label-1",
		"test-label-2",
	}

	ipfsMeta := nftv3.IPFSMetadata{
		Name:                  "test-name",
		Description:           "test-desc",
		Image:                 "test-image",
		DocumentAttributeKeys: docAttrKeyLabels,
	}

	mintReq := &nftv3.MintNFTRequest{
		DocumentID:      docID,
		CollectionID:    collectionID,
		Owner:           acc.GetIdentity(),
		IPFSMetadata:    ipfsMeta,
		GrantReadAccess: false,
	}

	mintRes, err := nftService.MintNFT(ctx, mintReq, true)
	assert.ErrorIs(t, err, nftv3.ErrDocumentRetrieval)
	assert.Nil(t, mintRes)
}

func TestIntegration_Service_MintNFT_NonPendingDocument_NonExistingCollection(t *testing.T) {
	acc, err := cfgService.GetAccount(keyrings.AliceKeyRingPair.PublicKey)
	assert.NoError(t, err)

	ctx := contextutil.WithAccount(context.Background(), acc)

	docID, err := createAndCommitGenericDocument(t, ctx, acc.GetIdentity())
	assert.NoError(t, err)

	collectionID := types.U64(rand.Int63())

	docAttrKeyLabels := []string{
		"test-label-1",
		"test-label-2",
	}

	ipfsMeta := nftv3.IPFSMetadata{
		Name:                  "test-name",
		Description:           "test-desc",
		Image:                 "test-image",
		DocumentAttributeKeys: docAttrKeyLabels,
	}

	mintReq := &nftv3.MintNFTRequest{
		DocumentID:      docID,
		CollectionID:    collectionID,
		Owner:           acc.GetIdentity(),
		IPFSMetadata:    ipfsMeta,
		GrantReadAccess: false,
	}

	mintRes, err := nftService.MintNFT(ctx, mintReq, false)
	assert.NoError(t, err)
	assert.NotNil(t, mintRes)

	jobID := hexutil.MustDecode(mintRes.JobID)
	result, err := dispatcher.Result(acc.GetIdentity(), jobID)
	assert.NoError(t, err)

	_, err = result.Await(ctx)
	assert.NoError(t, err)
}

func TestIntegration_Service_CreateNFTCollection(t *testing.T) {
	acc, err := cfgService.GetAccount(keyrings.AliceKeyRingPair.PublicKey)
	assert.NoError(t, err)

	ctx := contextutil.WithAccount(context.Background(), acc)
	collectionID := types.U64(1234)

	createCollectionReq := &nftv3.CreateNFTCollectionRequest{
		CollectionID: collectionID,
	}

	createCollectionRes, err := nftService.CreateNFTCollection(ctx, createCollectionReq)
	assert.NoError(t, err)
	assert.NotNil(t, createCollectionRes)

	jobID := gocelery.JobID(hexutil.MustDecode(createCollectionRes.JobID))
	result, err := dispatcher.Result(acc.GetIdentity(), jobID)
	assert.NoError(t, err)

	_, err = result.Await(ctx)
	assert.NoError(t, err)

	// Collection already exists
	createCollectionRes, err = nftService.CreateNFTCollection(ctx, createCollectionReq)
	assert.NotNil(t, err)
	assert.Nil(t, createCollectionRes)
}

func createAndCommitGenericDocument(t *testing.T, ctx context.Context, accountID *types.AccountID) ([]byte, error) {
	genericSrv, err := registry.LocateService(documenttypes.GenericDataTypeUrl)
	assert.NoError(t, err)

	attr1, err := documents.NewStringAttribute("test-label-1", documents.AttrString, "test-attribute-1")
	assert.NoError(t, err)

	attr2, err := documents.NewStringAttribute("test-label-2", documents.AttrInt256, "1234")
	assert.NoError(t, err)

	doc, err := genericSrv.Derive(
		ctx,
		documents.UpdatePayload{
			CreatePayload: documents.CreatePayload{
				Scheme: generic.Scheme,
				Collaborators: documents.CollaboratorsAccess{
					ReadWriteCollaborators: nil,
				},
				Attributes: map[documents.AttrKey]documents.Attribute{
					attr1.Key: attr1,
					attr2.Key: attr2,
				},
			},
		})
	assert.NoError(t, err)

	jobID, err := genericSrv.Commit(ctx, doc)
	assert.NoError(t, err)

	res, err := dispatcher.Result(accountID, jobID)
	assert.NoError(t, err)

	_, err = res.Await(ctx)
	assert.NoError(t, err)

	return doc.ID(), err
}

func createPendingDocument(t *testing.T, ctx context.Context) (documents.Document, error) {
	attr1, err := documents.NewStringAttribute("test-label-1", documents.AttrString, "test-attribute-1")
	assert.NoError(t, err)

	attr2, err := documents.NewStringAttribute("test-label-2", documents.AttrInt256, "1234")
	assert.NoError(t, err)

	doc, err := pendingDocSrv.Create(
		ctx,
		documents.UpdatePayload{
			CreatePayload: documents.CreatePayload{
				Scheme: generic.Scheme,
				Collaborators: documents.CollaboratorsAccess{
					ReadWriteCollaborators: nil,
				},
				Attributes: map[documents.AttrKey]documents.Attribute{
					attr1.Key: attr1,
					attr2.Key: attr2,
				},
			},
		})
	assert.NoError(t, err)

	return doc, err
}

func getGenericDocument(t *testing.T, ctx context.Context, docID []byte) (documents.Document, error) {
	genericSrv, err := registry.LocateService(documenttypes.GenericDataTypeUrl)
	assert.NoError(t, err)

	return genericSrv.GetCurrentVersion(ctx, docID)
}
