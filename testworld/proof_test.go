// +build testworld

package testworld

import (
	"net/http"
	"testing"

	"github.com/gavv/httpexpect"
)

func TestProofWithMultipleFields_invoice_successful(t *testing.T) {
	t.Parallel()
	proofWithMultipleFields_successful(t, typeDocuments)
}

func proofWithMultipleFields_successful(t *testing.T, documentType string) {
	alice := doctorFord.getHostTestSuite(t, "Alice")
	bob := doctorFord.getHostTestSuite(t, "Bob")

	// Alice shares a document with Bob
	res := createDocument(alice.httpExpect, alice.id.String(), documentType, http.StatusAccepted, genericCoreAPICreate([]string{bob.id.String()}))
	txID := getJobID(t, res)
	status, message := getTransactionStatusAndMessage(alice.httpExpect, alice.id.String(), txID)
	if status != "success" {
		t.Error(message)
	}

	docIdentifier := getDocumentIdentifier(t, res)
	if docIdentifier == "" {
		t.Error("docIdentifier empty")
	}

	proofPayload := defaultProofPayload(documentType)

	proofFromAlice := getProof(alice.httpExpect, alice.id.String(), http.StatusOK, docIdentifier, proofPayload)
	proofFromBob := getProof(bob.httpExpect, bob.id.String(), http.StatusOK, docIdentifier, proofPayload)

	checkProof(proofFromAlice, documentType, docIdentifier)
	checkProof(proofFromBob, documentType, docIdentifier)
}

func checkProof(objProof *httpexpect.Object, documentType string, docIdentifier string) {
	prop1 := "0x0005000000000001" // generic.Scheme
	prop2 := "0x0100000000000009" // cd_tree.documentIdentifier
	objProof.Path("$.header.document_id").String().Equal(docIdentifier)
	objProof.Path("$.field_proofs[0].property").String().Equal(prop1)
	objProof.Path("$.field_proofs[0].sorted_hashes").NotNull()
	objProof.Path("$.field_proofs[1].property").String().Equal(prop2)
	objProof.Path("$.field_proofs[1].sorted_hashes").NotNull()
}
