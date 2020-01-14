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
	res := createDocument(alice.httpExpect, alice.id.String(), documentType, http.StatusAccepted, defaultDocumentPayload(documentType, []string{bob.id.String()}))
	txID := getTransactionID(t, res)
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
	compactPrefix := "0x00010000" // invoice prefix
	prop1 := "0000002d"           // invoice.net_amount
	prop2 := "0000000d"           // invoice.currency
	objProof.Path("$.header.document_id").String().Equal(docIdentifier)
	objProof.Path("$.field_proofs[0].property").String().Equal(compactPrefix + prop1)
	objProof.Path("$.field_proofs[0].sorted_hashes").NotNull()
	objProof.Path("$.field_proofs[1].property").String().Equal(compactPrefix + prop2)
	objProof.Path("$.field_proofs[1].sorted_hashes").NotNull()
}
