// +build testworld

package testworld

import (
	"net/http"
	"testing"

	"github.com/gavv/httpexpect"
)

func TestProofWithMultipleFields_successful(t *testing.T) {
	alice := doctorFord.getHostTestSuite(t, "Alice")
	bob := doctorFord.getHostTestSuite(t, "Bob")

	// Alice shares invoice document with Bob
	res, err := alice.host.createInvoice(alice.httpExpect, http.StatusOK, defaultNFTPayload([]string{bob.id.String()}))
	if err != nil {
		t.Error(err)
	}

	docIdentifier := getDocumentIdentifier(t, res)
	if docIdentifier == "" {
		t.Error("docIdentifier empty")
	}

	proofPayload := defaultProofPayload()

	proofFromAlice := getProof(alice.httpExpect, http.StatusOK, docIdentifier, proofPayload)
	proofFromBob := getProof(bob.httpExpect, http.StatusOK, docIdentifier, proofPayload)

	checkProof(proofFromAlice, docIdentifier)
	checkProof(proofFromBob, docIdentifier)

}

func checkProof(objProof *httpexpect.Object, docIdentifier string) {
	objProof.Path("$.header.document_id").String().Equal(docIdentifier)
	objProof.Path("$.field_proofs[0].property").String().Equal("invoice.net_amount")
	objProof.Path("$.field_proofs[0].sorted_hashes").NotNull()
	objProof.Path("$.field_proofs[1].property").String().Equal("invoice.currency")
	objProof.Path("$.field_proofs[1].sorted_hashes").NotNull()

}
