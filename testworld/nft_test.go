//go:build testworld

package testworld

import (
	"encoding/json"
	"fmt"
	"math/big"
	"math/rand"
	"net/http"
	"testing"
	"time"

	proxyType "github.com/centrifuge/chain-custom-types/pkg/proxy"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/http/coreapi"
	"github.com/centrifuge/go-centrifuge/ipfs"
	nftv3 "github.com/centrifuge/go-centrifuge/nft/v3"
	"github.com/centrifuge/go-centrifuge/testworld/park/behavior/client"
	"github.com/centrifuge/go-centrifuge/testworld/park/host"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ipfs/go-cid"
	"github.com/ipfs/interface-go-ipfs-core/path"
	mh "github.com/multiformats/go-multihash"
	"github.com/stretchr/testify/assert"
)

func TestNFTAPI_Mint_CommitEnabled(t *testing.T) {
	t.Parallel()

	charlie, err := controller.GetHost(host.Alice)
	assert.NoError(t, err)
	bob, err := controller.GetHost(host.Bob)
	assert.NoError(t, err)

	charlieJWT, err := charlie.GetMainAccount().GetJW3Token(proxyType.ProxyTypeName[proxyType.PodAuth])
	assert.NoError(t, err)

	charlieClient := client.New(t, controller.GetWebhookReceiver(), charlie.GetAPIURL(), charlieJWT)

	docPayload := genericCoreAPICreate([]string{
		charlie.GetMainAccount().GetAccountID().ToHexString(),
		bob.GetMainAccount().GetAccountID().ToHexString(),
	})

	attrs, _ := getAttributeMapRequest(t, charlie.GetMainAccount().GetAccountID())
	docPayload["attributes"] = attrs

	res := charlieClient.CreateDocument(
		"documents",
		http.StatusCreated,
		docPayload,
	)
	assert.Equal(t, client.GetDocumentStatus(res), "pending")
	docID := client.GetDocumentIdentifier(res)

	collectionID := types.U64(rand.Int63())

	payload := map[string]interface{}{
		"collection_id": collectionID,
	}

	createClassRes := charlieClient.CreateNFTCollection(http.StatusAccepted, payload)

	jobID, err := client.GetJobID(createClassRes)
	assert.NoError(t, err)

	err = charlieClient.WaitForJobCompletion(jobID)
	assert.NoError(t, err)

	// Use the same attributes that were used when the doc was created.

	var docAttrs []string
	docAttrMap := make(map[string]string)

	for attr, req := range attrs {
		docAttrs = append(docAttrs, attr)
		docAttrMap[attr] = req.Value
	}

	ipfsName := "ipfs_name"
	ipfsDescription := "ipfs_description"
	ipfsImage := "ipfs_image"

	ipfsMetadata := nftv3.IPFSMetadata{
		Name:                  ipfsName,
		Description:           ipfsDescription,
		Image:                 ipfsImage,
		DocumentAttributeKeys: docAttrs,
	}

	payload = map[string]interface{}{
		"collection_id":   collectionID,
		"document_id":     docID,
		"owner":           charlie.GetMainAccount().GetAccountID().ToHexString(),
		"ipfs_metadata":   ipfsMetadata,
		"freeze_metadata": false,
	}

	mintRes := charlieClient.CommitAndMintNFT(http.StatusAccepted, payload)

	jobID, err = client.GetJobID(mintRes)
	assert.NoError(t, err)

	err = charlieClient.WaitForJobCompletion(jobID)
	assert.NoError(t, err)

	docVal := charlieClient.GetDocumentAndVerify(docID, nil, attrs)
	itemIDRaw := docVal.Path("$.header.nfts[0].item_id").String().Raw()

	i := new(big.Int)
	bi, ok := i.SetString(itemIDRaw, 10)
	assert.True(t, ok)

	itemID := types.NewU128(*bi)

	mintOwner := mintRes.Value("owner").String().Raw()
	assert.NotEmpty(t, mintOwner, "mint owner is empty")

	payload = map[string]interface{}{
		"collection_id": collectionID,
		"item_id":       itemID,
	}

	ownerRes := charlieClient.GetOwnerOfNFT(http.StatusOK, payload)

	resOwner := ownerRes.Value("owner").String().Raw()
	assert.Equal(t, mintOwner, resOwner, "owners should be equal")

	payload = map[string]interface{}{
		"collection_id": collectionID,
		"item_id":       itemID,
	}

	metadataRes := charlieClient.GetMetadataOfNFT(http.StatusOK, payload)

	nftMetadata := ipfs.NFTMetadata{
		Name:        ipfsName,
		Description: ipfsDescription,
		Image:       ipfsImage,
		Properties:  docAttrMap,
	}

	nftMetadataJSONBytes, err := json.Marshal(nftMetadata)
	assert.NoError(t, err)

	v1CidPrefix := cid.Prefix{
		Codec:    cid.Raw,
		MhLength: -1,
		MhType:   mh.SHA2_256,
		Version:  1,
	}

	metadataCID, err := v1CidPrefix.Sum(nftMetadataJSONBytes)
	assert.NoError(t, err)

	metaPath := path.New(metadataCID.String())

	resData := metadataRes.Value("data").String().Raw()

	decodedResData, err := hexutil.Decode(resData)
	assert.NoError(t, err)

	assert.Equal(t, metaPath.String(), string(decodedResData))

	resFrozen := metadataRes.Value("is_frozen").Boolean().Raw()
	assert.False(t, resFrozen)

	payload = map[string]interface{}{
		"collection_id":  collectionID,
		"item_id":        itemID,
		"attribute_name": nftv3.DocumentIDAttributeKey,
	}

	docIDAttributeRes := charlieClient.GetAttributeOfNFT(http.StatusOK, payload)

	resDocumentID := docIDAttributeRes.Value("value").String().Raw()

	assert.Equal(t, docID, resDocumentID)

	docVersion := docVal.Path("$.header.version_id").String().Raw()

	payload = map[string]interface{}{
		"collection_id":  collectionID,
		"item_id":        itemID,
		"attribute_name": nftv3.DocumentVersionAttributeKey,
	}

	docVersionAttributeRes := charlieClient.GetAttributeOfNFT(http.StatusOK, payload)

	resDocumentVersion := docVersionAttributeRes.Value("value").String().Raw()

	assert.Equal(t, docVersion, resDocumentVersion)
}

func TestNFTAPI_Mint_CommitDisabled(t *testing.T) {
	t.Parallel()

	bob, err := controller.GetHost(host.Bob)
	assert.NoError(t, err)

	charlie, err := controller.GetHost(host.Charlie)
	assert.NoError(t, err)

	bobJWT, err := bob.GetMainAccount().GetJW3Token(proxyType.ProxyTypeName[proxyType.PodAuth])
	assert.NoError(t, err)

	charlieJWT, err := charlie.GetMainAccount().GetJW3Token(proxyType.ProxyTypeName[proxyType.PodAuth])
	assert.NoError(t, err)

	bobClient := client.New(t, controller.GetWebhookReceiver(), bob.GetAPIURL(), bobJWT)
	charlieClient := client.New(t, controller.GetWebhookReceiver(), charlie.GetAPIURL(), charlieJWT)

	// Bob shares document with Charlie
	docPayload := genericCoreAPICreate([]string{
		bob.GetMainAccount().GetAccountID().ToHexString(),
		charlie.GetMainAccount().GetAccountID().ToHexString(),
	})

	attrs, _ := getAttributeMapRequest(t, bob.GetMainAccount().GetAccountID())
	docPayload["attributes"] = attrs

	docID, err := bobClient.CreateAndCommitDocument(docPayload)
	assert.NoError(t, err)

	bobClient.GetDocumentAndVerify(docID, nil, attrs)
	charlieClient.GetDocumentAndVerify(docID, nil, attrs)

	collectionID := types.U64(rand.Int63())

	payload := map[string]interface{}{
		"collection_id": collectionID,
	}

	createClassRes := bobClient.CreateNFTCollection(http.StatusAccepted, payload)

	jobID, err := client.GetJobID(createClassRes)
	assert.NoError(t, err)

	err = bobClient.WaitForJobCompletion(jobID)
	assert.NoError(t, err)

	// Use the same attributes that were used when the doc was created.

	var docAttrs []string
	docAttrMap := make(map[string]string)

	for attr, req := range attrs {
		docAttrs = append(docAttrs, attr)
		docAttrMap[attr] = req.Value
	}

	ipfsName := "ipfs_name"
	ipfsDescription := "ipfs_description"
	ipfsImage := "ipfs_image"

	ipfsMetadata := nftv3.IPFSMetadata{
		Name:                  ipfsName,
		Description:           ipfsDescription,
		Image:                 ipfsImage,
		DocumentAttributeKeys: docAttrs,
	}

	payload = map[string]interface{}{
		"collection_id":   collectionID,
		"document_id":     docID,
		"owner":           bob.GetMainAccount().GetAccountID().ToHexString(),
		"ipfs_metadata":   ipfsMetadata,
		"freeze_metadata": false,
	}

	mintRes := bobClient.MintNFT(http.StatusAccepted, payload)

	jobID, err = client.GetJobID(mintRes)
	assert.NoError(t, err)

	err = bobClient.WaitForJobCompletion(jobID)
	assert.NoError(t, err)

	docVal := bobClient.GetDocumentAndVerify(docID, nil, attrs)
	itemIDRaw := docVal.Path("$.header.nfts[0].item_id").String().Raw()

	i := new(big.Int)
	bi, ok := i.SetString(itemIDRaw, 10)
	assert.True(t, ok)

	itemID := types.NewU128(*bi)

	mintOwner := mintRes.Value("owner").String().Raw()
	assert.NotEmpty(t, mintOwner, "mint owner is empty")

	payload = map[string]interface{}{
		"collection_id": collectionID,
		"item_id":       itemID,
	}

	ownerRes := bobClient.GetOwnerOfNFT(http.StatusOK, payload)

	resOwner := ownerRes.Value("owner").String().Raw()
	assert.Equal(t, mintOwner, resOwner, "owners should be equal")

	payload = map[string]interface{}{
		"collection_id": collectionID,
		"item_id":       itemID,
	}

	metadataRes := bobClient.GetMetadataOfNFT(http.StatusOK, payload)

	nftMetadata := ipfs.NFTMetadata{
		Name:        ipfsName,
		Description: ipfsDescription,
		Image:       ipfsImage,
		Properties:  docAttrMap,
	}

	nftMetadataJSONBytes, err := json.Marshal(nftMetadata)
	assert.NoError(t, err)

	v1CidPrefix := cid.Prefix{
		Codec:    cid.Raw,
		MhLength: -1,
		MhType:   mh.SHA2_256,
		Version:  1,
	}

	metadataCID, err := v1CidPrefix.Sum(nftMetadataJSONBytes)
	assert.NoError(t, err)

	metaPath := path.New(metadataCID.String())

	resData := metadataRes.Value("data").String().Raw()

	decodedResData, err := hexutil.Decode(resData)
	assert.NoError(t, err)

	assert.Equal(t, metaPath.String(), string(decodedResData))

	resFrozen := metadataRes.Value("is_frozen").Boolean().Raw()
	assert.False(t, resFrozen)

	payload = map[string]interface{}{
		"collection_id":  collectionID,
		"item_id":        itemID,
		"attribute_name": nftv3.DocumentIDAttributeKey,
	}

	docIDAttributeRes := bobClient.GetAttributeOfNFT(http.StatusOK, payload)

	resDocumentID := docIDAttributeRes.Value("value").String().Raw()

	assert.Equal(t, docID, resDocumentID)

	docVersion := docVal.Path("$.header.version_id").String().Raw()

	payload = map[string]interface{}{
		"collection_id":  collectionID,
		"item_id":        itemID,
		"attribute_name": nftv3.DocumentVersionAttributeKey,
	}

	docVersionAttributeRes := bobClient.GetAttributeOfNFT(http.StatusOK, payload)

	resDocumentVersion := docVersionAttributeRes.Value("value").String().Raw()

	assert.Equal(t, docVersion, resDocumentVersion)
}

func getAttributeMapRequest(t *testing.T, identity *types.AccountID) (coreapi.AttributeMapRequest, []string) {
	attrs, pfs := getAttributes(t, identity)
	amr := make(coreapi.AttributeMapRequest)
	for _, attr := range attrs {
		v, err := attr.Value.String()
		assert.NoError(t, err)
		amr[attr.KeyLabel] = coreapi.AttributeRequest{
			Type:  attr.Value.Type.String(),
			Value: v,
		}
	}
	return amr, pfs
}

func getAttributes(t *testing.T, identity *types.AccountID) (map[documents.AttrKey]documents.Attribute, []string) {
	attrs := map[documents.AttrKey]documents.Attribute{}
	attr1, err := documents.NewStringAttribute("Originator", documents.AttrBytes, identity.ToHexString())
	assert.NoError(t, err)
	attrs[attr1.Key] = attr1
	attr2, err := documents.NewStringAttribute("AssetValue", documents.AttrDecimal, "100")
	assert.NoError(t, err)
	attrs[attr2.Key] = attr2
	attr3, err := documents.NewStringAttribute("AssetIdentifier", documents.AttrBytes, hexutil.Encode(utils.RandomSlice(32)))
	assert.NoError(t, err)
	attrs[attr3.Key] = attr3
	attr4, err := documents.NewStringAttribute("MaturityDate", documents.AttrTimestamp, time.Now().Format(time.RFC3339Nano))
	assert.NoError(t, err)
	attrs[attr4.Key] = attr4
	attr5, err := documents.NewStringAttribute("result", documents.AttrBytes,
		hexutil.Encode([]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 100}))
	assert.NoError(t, err)
	attrs[attr5.Key] = attr5
	var proofFields []string
	for _, a := range []documents.Attribute{attr1, attr2, attr3, attr4} {
		proofFields = append(proofFields, fmt.Sprintf("%s.attributes[%s].byte_val", documents.CDTreePrefix, a.Key.String()))
	}
	return attrs, proofFields
}
