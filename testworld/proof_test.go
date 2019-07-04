// +build testworld

package testworld

import (
	"net/http"
	"testing"

	"github.com/gavv/httpexpect"
)

func TestProofWithMultipleFields_invoice_successful(t *testing.T) {
	t.Parallel()
	proofWithMultipleFields_successful(t, typeInvoice)
}

func TestProofWithMultipleFields_po_successful(t *testing.T) {
	t.Parallel()
	proofWithMultipleFields_successful(t, typePO)
}

func proofWithMultipleFields_successful(t *testing.T, documentType string) {
	alice := doctorFord.getHostTestSuite(t, "Alice")
	bob := doctorFord.getHostTestSuite(t, "Bob")

	st := http.StatusOK
	if documentType == typePO {
		// TODO(ved): remove once invoice is moved to new apis
		st = http.StatusCreated
	}

	// Alice shares a document with Bob
	res := createDocument(alice.httpExpect, alice.id.String(), documentType, st, defaultDocumentPayload(documentType, []string{bob.id.String()}))
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

	if documentType == typePO {
		compactPrefix = "0x00020000" // po prefix
		prop1 = "00000012"           // po.total_amount
		prop2 = "00000011"           // po.currency
	}

	objProof.Path("$.header.document_id").String().Equal(docIdentifier)
	objProof.Path("$.field_proofs[0].property").String().Equal(compactPrefix + prop1)
	objProof.Path("$.field_proofs[0].sorted_hashes").NotNull()
	objProof.Path("$.field_proofs[1].property").String().Equal(compactPrefix + prop2)
	objProof.Path("$.field_proofs[1].sorted_hashes").NotNull()
}
