package testworld

import (
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

	clasIDHex, err := types.EncodeToHexString(classID)
	assert.NoError(t, err, "encoding the class ID to hex should be successful")

	payload := map[string]interface{}{
		"class_id":    clasIDHex,
		"document_id": docID,
		"public_info": []string{"test"},
		"owner":       alice.host.centChainID,
	}

	mintRes, err := alice.host.mintNFTV3(alice.httpExpect, alice.id.String(), http.StatusAccepted, payload)
	assert.NoError(t, err, "mintNFTV3 should be successful")
	jobID := getJobID(t, mintRes)
	err = waitForJobComplete(doctorFord.maeve, alice.httpExpect, alice.id.String(), jobID)
	assert.NoError(t, err)

	docVal := getDocumentAndVerify(t, alice.httpExpect, alice.id.String(), docID, nil, attrs)

	instanceIDhex := docVal.Path("$.header.cc_nfts[0].instance_id").String().Raw()

	mintOwner := mintRes.Value("owner").String().Raw()
	assert.NotEmpty(t, mintOwner, "mint owner is empty")

	payload = map[string]interface{}{
		"class_id":    clasIDHex,
		"instance_id": instanceIDhex,
	}

	ownerRes, err := alice.host.ownerOfNFTV3(alice.httpExpect, alice.id.String(), http.StatusOK, payload)
	assert.NoError(t, err, "ownerOfNFTV3 should be successful")

	resOwner := ownerRes.Value("owner").String().Raw()
	assert.Equal(t, mintOwner, resOwner, "owners should be equal")
}
