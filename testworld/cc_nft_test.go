//go:build testworld
// +build testworld

package testworld

import (
	"encoding/json"
	"math/big"
	"net/http"
	"testing"

	mh "github.com/multiformats/go-multihash"

	nftv3 "github.com/centrifuge/go-centrifuge/nft/v3"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ipfs/go-cid"
	"github.com/ipfs/interface-go-ipfs-core/path"
	"github.com/stretchr/testify/assert"
)

func TestCcNFTMint(t *testing.T) {
	alice := doctorFord.getHostTestSuite(t, "Alice")
	bob := doctorFord.getHostTestSuite(t, "Bob")

	// Alice shares document with Bob
	docPayload := genericCoreAPICreate([]string{alice.id.String(), bob.id.String()})
	attrs, _ := getAttributeMapRequest(t, alice.id)
	docPayload["attributes"] = attrs
	docID := createAndCommitDocument(t, doctorFord.maeve, alice.httpExpect, alice.id.String(), docPayload)
	getDocumentAndVerify(t, alice.httpExpect, alice.id.String(), docID, nil, attrs)
	getDocumentAndVerify(t, bob.httpExpect, bob.id.String(), docID, nil, attrs)

	classID := types.U64(1234)

	payload := map[string]interface{}{
		"class_id": classID,
	}

	createClassRes, err := alice.host.createNFTClassV3(alice.httpExpect, alice.id.String(), http.StatusAccepted, payload)
	assert.NoError(t, err, "createNFTClassV3 should be successful")

	jobID := getJobID(t, createClassRes)
	err = waitForJobComplete(doctorFord.maeve, alice.httpExpect, alice.id.String(), jobID)
	assert.NoError(t, err)

	// Use the same attributes that were used when the doc was created.

	var docAttrs []string
	docAttrMap := make(map[string]string)

	for attr, req := range attrs {
		docAttrs = append(docAttrs, attr)
		docAttrMap[attr] = req.Value
	}

	payload = map[string]interface{}{
		"class_id":            classID,
		"document_id":         docID,
		"owner":               alice.host.cfgVals.CentChainID,
		"document_attributes": docAttrs,
		"freeze_metadata":     true,
	}

	mintRes, err := alice.host.mintNFTV3(alice.httpExpect, alice.id.String(), http.StatusAccepted, payload)
	assert.NoError(t, err, "mintNFTV3 should be successful")
	jobID = getJobID(t, mintRes)
	err = waitForJobComplete(doctorFord.maeve, alice.httpExpect, alice.id.String(), jobID)
	assert.NoError(t, err)

	docVal := getDocumentAndVerify(t, alice.httpExpect, alice.id.String(), docID, nil, attrs)

	instanceIDraw := docVal.Path("$.header.cc_nfts[0].instance_id").String().Raw()

	i := new(big.Int)
	bi, ok := i.SetString(instanceIDraw, 10)
	assert.True(t, ok)

	instanceID := types.NewU128(*bi)

	mintOwner := mintRes.Value("owner").String().Raw()
	assert.NotEmpty(t, mintOwner, "mint owner is empty")

	payload = map[string]interface{}{
		"class_id":    classID,
		"instance_id": instanceID.String(),
	}

	ownerRes, err := alice.host.ownerOfNFTV3(alice.httpExpect, alice.id.String(), http.StatusOK, payload)
	assert.NoError(t, err, "ownerOfNFTV3 should be successful")

	resOwner := ownerRes.Value("owner").String().Raw()
	assert.Equal(t, mintOwner, resOwner, "owners should be equal")

	payload = map[string]interface{}{
		"class_id":    classID,
		"instance_id": instanceID.String(),
	}

	metadataRes, err := alice.host.metadataOfNFTV3(alice.httpExpect, alice.id.String(), http.StatusOK, payload)
	assert.NoError(t, err, "ownerOfNFTV3 should be successful")

	docIDBytes, err := hexutil.Decode(docID)
	assert.NoError(t, err)

	docVersionBytes, err := hexutil.Decode(docVal.Path("$.header.version_id").String().Raw())
	assert.NoError(t, err)

	nftMetadata := nftv3.NFTMetadata{
		DocID:         docIDBytes,
		DocVersion:    docVersionBytes,
		DocAttributes: docAttrMap,
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
	assert.True(t, resFrozen)
}
