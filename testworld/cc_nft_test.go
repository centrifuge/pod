//go:build testworld
// +build testworld

package testworld

import (
	"math/big"
	"net/http"
	"testing"

	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
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

	nftMetadata := "nft_metadata"

	payload = map[string]interface{}{
		"class_id":        classID,
		"document_id":     docID,
		"owner":           alice.host.centChainID,
		"metadata":        nftMetadata,
		"freeze_metadata": true,
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

	resMeta := metadataRes.Value("data").String().Raw()
	assert.Equal(t, nftMetadata, resMeta)

	resFrozen := metadataRes.Value("is_frozen").Boolean().Raw()
	assert.True(t, resFrozen)
}
