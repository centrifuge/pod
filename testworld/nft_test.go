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

	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/http/coreapi"
	nftv3 "github.com/centrifuge/go-centrifuge/nft/v3"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ipfs/go-cid"
	"github.com/ipfs/interface-go-ipfs-core/path"
	mh "github.com/multiformats/go-multihash"
	"github.com/stretchr/testify/assert"
)

func TestCcNFTMint_CommitDisabled(t *testing.T) {
	alice := doctorFord.getHostTestSuite(t, testAccountAlice)
	bob := doctorFord.getHostTestSuite(t, testAccountBob)

	aliceAccountID, err := alice.testAccount.AccountID()
	assert.NoError(t, err)

	bobAccountID, err := bob.testAccount.AccountID()
	assert.NoError(t, err)

	aliceJW3T, err := alice.testAccount.toMockJW3T()
	assert.NoError(t, err)

	bobJW3T, err := bob.testAccount.toMockJW3T()
	assert.NoError(t, err)

	// Alice shares document with Bob
	docPayload := genericCoreAPICreate([]string{
		aliceAccountID.ToHexString(),
		bobAccountID.ToHexString(),
	})

	attrs, _ := getAttributeMapRequest(t, aliceAccountID)
	docPayload["attributes"] = attrs
	docID := createAndCommitDocument(
		t,
		doctorFord.maeve,
		alice.httpExpect,
		aliceJW3T,
		docPayload,
	)

	getDocumentAndVerify(t, alice.httpExpect, aliceJW3T, docID, nil, attrs)
	getDocumentAndVerify(t, bob.httpExpect, bobJW3T, docID, nil, attrs)

	collectionID := types.U64(rand.Int63())

	payload := map[string]interface{}{
		"collection_id": collectionID,
	}

	createClassRes := createNFTCollectionV3(alice.httpExpect, aliceJW3T, http.StatusAccepted, payload)

	jobID := getJobID(t, createClassRes)
	err = waitForJobComplete(doctorFord.maeve, alice.httpExpect, aliceJW3T, jobID)
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
		"owner":           aliceAccountID.ToHexString(),
		"ipfs_metadata":   ipfsMetadata,
		"freeze_metadata": false,
	}

	mintRes := mintNFTV3(alice.httpExpect, aliceJW3T, http.StatusAccepted, payload)

	jobID = getJobID(t, mintRes)
	err = waitForJobComplete(doctorFord.maeve, alice.httpExpect, aliceJW3T, jobID)
	assert.NoError(t, err)
	docVal := getDocumentAndVerify(t, alice.httpExpect, aliceJW3T, docID, nil, attrs)
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

	ownerRes := ownerOfNFTV3(alice.httpExpect, aliceJW3T, http.StatusOK, payload)

	resOwner := ownerRes.Value("owner").String().Raw()
	assert.Equal(t, mintOwner, resOwner, "owners should be equal")

	payload = map[string]interface{}{
		"collection_id": collectionID,
		"item_id":       itemID,
	}

	metadataRes := metadataOfNFTV3(alice.httpExpect, aliceJW3T, http.StatusOK, payload)

	nftMetadata := nftv3.NFTMetadata{
		Name:        ipfsName,
		Description: ipfsDescription,
		Image:       ipfsImage,
		Properties:  docAttrMap,
	}

	nftMetadataJSONBytes, err := json.Marshal(nftMetadata)
	assert.NoError(t, err)

	var v1CidPrefix = cid.Prefix{
		Codec:    cid.Raw,
		MhLength: -1,
		MhType:   mh.SHA2_256,
		Version:  1,
	}

	metadataCID, err := v1CidPrefix.Sum(nftMetadataJSONBytes)
	assert.NoError(t, err)

	metaPath := path.New(metadataCID.String())

	resData := metadataRes.Value("data").String().Raw()

	assert.Equal(t, metaPath.String(), resData)

	resFrozen := metadataRes.Value("is_frozen").Boolean().Raw()
	assert.False(t, resFrozen)

	payload = map[string]interface{}{
		"collection_id":  collectionID,
		"item_id":        itemID,
		"attribute_name": nftv3.DocumentIDAttributeKey,
	}

	docIDAttributeRes := attributeOfNFTV3(alice.httpExpect, aliceJW3T, http.StatusOK, payload)

	resDocumentID := docIDAttributeRes.Value("value").String().Raw()

	assert.Equal(t, docID, resDocumentID)

	docVersion := docVal.Path("$.header.version_id").String().Raw()

	payload = map[string]interface{}{
		"collection_id":  collectionID,
		"item_id":        itemID,
		"attribute_name": nftv3.DocumentVersionAttributeKey,
	}

	docVersionAttributeRes := attributeOfNFTV3(alice.httpExpect, aliceJW3T, http.StatusOK, payload)

	resDocumentVersion := docVersionAttributeRes.Value("value").String().Raw()

	assert.Equal(t, docVersion, resDocumentVersion)
}

func TestCcNFTMint_CommitEnabled(t *testing.T) {
	alice := doctorFord.getHostTestSuite(t, testAccountAlice)
	bob := doctorFord.getHostTestSuite(t, testAccountBob)

	aliceAccountID, err := alice.testAccount.AccountID()
	assert.NoError(t, err)

	bobAccountID, err := bob.testAccount.AccountID()
	assert.NoError(t, err)

	aliceJW3T, err := alice.testAccount.toMockJW3T()
	assert.NoError(t, err)

	// Alice shares document with Bob
	docPayload := genericCoreAPICreate([]string{
		aliceAccountID.ToHexString(),
		bobAccountID.ToHexString(),
	})

	attrs, _ := getAttributeMapRequest(t, aliceAccountID)
	docPayload["attributes"] = attrs
	res := createDocument(
		alice.httpExpect,
		aliceJW3T,
		"documents",
		http.StatusCreated,
		docPayload,
	)
	status := getDocumentStatus(t, res)
	assert.Equal(t, status, "pending")
	docID := getDocumentIdentifier(t, res)

	// TODO(cdamian): Enable these
	//getDocumentAndVerify(t, alice.httpExpect, aliceJW3T, docID, nil, attrs)
	//getDocumentAndVerify(t, bob.httpExpect, bobJW3T, docID, nil, attrs)

	collectionID := types.U64(rand.Int63())

	payload := map[string]interface{}{
		"collection_id": collectionID,
	}

	createClassRes := createNFTCollectionV3(alice.httpExpect, aliceJW3T, http.StatusAccepted, payload)

	jobID := getJobID(t, createClassRes)
	err = waitForJobComplete(doctorFord.maeve, alice.httpExpect, aliceJW3T, jobID)
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
		"owner":           aliceAccountID.ToHexString(),
		"ipfs_metadata":   ipfsMetadata,
		"freeze_metadata": false,
	}

	mintRes := commitAndMintNFTV3(alice.httpExpect, aliceJW3T, http.StatusAccepted, payload)

	jobID = getJobID(t, mintRes)
	err = waitForJobComplete(doctorFord.maeve, alice.httpExpect, aliceJW3T, jobID)
	assert.NoError(t, err)
	docVal := getDocumentAndVerify(t, alice.httpExpect, aliceJW3T, docID, nil, attrs)
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

	ownerRes := ownerOfNFTV3(alice.httpExpect, aliceJW3T, http.StatusOK, payload)

	resOwner := ownerRes.Value("owner").String().Raw()
	assert.Equal(t, mintOwner, resOwner, "owners should be equal")

	payload = map[string]interface{}{
		"collection_id": collectionID,
		"item_id":       itemID,
	}

	metadataRes := metadataOfNFTV3(alice.httpExpect, aliceJW3T, http.StatusOK, payload)

	nftMetadata := nftv3.NFTMetadata{
		Name:        ipfsName,
		Description: ipfsDescription,
		Image:       ipfsImage,
		Properties:  docAttrMap,
	}

	nftMetadataJSONBytes, err := json.Marshal(nftMetadata)
	assert.NoError(t, err)

	var v1CidPrefix = cid.Prefix{
		Codec:    cid.Raw,
		MhLength: -1,
		MhType:   mh.SHA2_256,
		Version:  1,
	}

	metadataCID, err := v1CidPrefix.Sum(nftMetadataJSONBytes)
	assert.NoError(t, err)

	metaPath := path.New(metadataCID.String())

	resData := metadataRes.Value("data").String().Raw()

	assert.Equal(t, metaPath.String(), resData)

	resFrozen := metadataRes.Value("is_frozen").Boolean().Raw()
	assert.False(t, resFrozen)

	payload = map[string]interface{}{
		"collection_id":  collectionID,
		"item_id":        itemID,
		"attribute_name": nftv3.DocumentIDAttributeKey,
	}

	docIDAttributeRes := attributeOfNFTV3(alice.httpExpect, aliceJW3T, http.StatusOK, payload)

	resDocumentID := docIDAttributeRes.Value("value").String().Raw()

	assert.Equal(t, docID, resDocumentID)

	docVersion := docVal.Path("$.header.version_id").String().Raw()

	payload = map[string]interface{}{
		"collection_id":  collectionID,
		"item_id":        itemID,
		"attribute_name": nftv3.DocumentVersionAttributeKey,
	}

	docVersionAttributeRes := attributeOfNFTV3(alice.httpExpect, aliceJW3T, http.StatusOK, payload)

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
