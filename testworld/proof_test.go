//go:build testworld

package testworld

import (
	"net/http"
	"testing"

	"github.com/centrifuge/pod/testworld/park/host"
	"github.com/gavv/httpexpect"
	"github.com/stretchr/testify/assert"
)

func TestDocumentsAPI_Proofs(t *testing.T) {
	t.Parallel()

	bob, err := controller.GetHost(host.Bob)
	assert.NoError(t, err)

	aliceClient, err := controller.GetClientForHost(t, host.Alice)
	assert.NoError(t, err)
	bobClient, err := controller.GetClientForHost(t, host.Bob)
	assert.NoError(t, err)

	// Alice shares a document with Bob
	payload := genericCoreAPICreate([]string{bob.GetMainAccount().GetAccountID().ToHexString()})

	docID, err := aliceClient.CreateAndCommitDocument(payload)
	assert.NoError(t, err)

	proofPayload := defaultProofPayload("documents")

	proofFromAlice := aliceClient.GetProof(http.StatusOK, docID, proofPayload)
	proofFromBob := bobClient.GetProof(http.StatusOK, docID, proofPayload)

	checkProof(proofFromAlice, docID)
	checkProof(proofFromBob, docID)
}

func checkProof(objProof *httpexpect.Object, docIdentifier string) {
	prop1 := "0x0005000000000001" // generic.Scheme
	prop2 := "0x0100000000000009" // cd_tree.documentIdentifier
	objProof.Path("$.header.document_id").String().Equal(docIdentifier)
	objProof.Path("$.field_proofs[0].property").String().Equal(prop1)
	objProof.Path("$.field_proofs[0].sorted_hashes").NotNull()
	objProof.Path("$.field_proofs[1].property").String().Equal(prop2)
	objProof.Path("$.field_proofs[1].sorted_hashes").NotNull()
}
