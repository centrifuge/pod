// +build testworld

package testworld

import (
	"net/http"
	"testing"

	"github.com/gavv/httpexpect"
)

func TestProofWithMultipleFields_invoice_successful(t *testing.T) {
	t.Parallel()
	proofWithMultipleFields_successful(t, TypeInvoice)

}

func TestProofWithMultipleFields_po_successful(t *testing.T) {
	t.Parallel()
	proofWithMultipleFields_successful(t, TypePO)

}

func proofWithMultipleFields_successful(t *testing.T, documentType string) {
	alice := doctorFord.getHostTestSuite(t, "Alice")
	bob := doctorFord.getHostTestSuite(t, "Bob")

	// Alice shares a document with Bob
	res := createDocument(alice.httpExpect, documentType, http.StatusOK, defaultDocumentPayload(documentType, []string{bob.id.String()}))

	docIdentifier := getDocumentIdentifier(t, res)
	if docIdentifier == "" {
		t.Error("docIdentifier empty")
	}

	proofPayload := defaultProofPayload(documentType)

	proofFromAlice := getProof(alice.httpExpect, http.StatusOK, docIdentifier, proofPayload)
	proofFromBob := getProof(bob.httpExpect, http.StatusOK, docIdentifier, proofPayload)

	checkProof(proofFromAlice, documentType, docIdentifier)
	checkProof(proofFromBob, documentType, docIdentifier)

}

func checkProof(objProof *httpexpect.Object, documentType string, docIdentifier string) {

	if documentType == TypePO {
		documentType = POPrefix
	}

	objProof.Path("$.header.document_id").String().Equal(docIdentifier)
	objProof.Path("$.field_proofs[0].property").String().Equal(documentType + ".net_amount")
	objProof.Path("$.field_proofs[0].sorted_hashes").NotNull()
	objProof.Path("$.field_proofs[1].property").String().Equal(documentType + ".currency")
	objProof.Path("$.field_proofs[1].sorted_hashes").NotNull()

}
