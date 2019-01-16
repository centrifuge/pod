// +build testworld

package testworld

import (
	"math/rand"
	"net/http"
	"testing"
	"time"

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
	// TODO remove this when we have retry for tasks
	time.Sleep(time.Duration(rand.Intn(5)) * time.Second)
	alice := doctorFord.getHostTestSuite(t, "Alice")
	bob := doctorFord.getHostTestSuite(t, "Bob")

	// Alice shares a document with Bob
	res := createDocument(alice.httpExpect, alice.id.String(), documentType, http.StatusOK, defaultDocumentPayload(documentType, []string{bob.id.String()}))
	txID := getTransactionID(t, res)
	waitTillStatus(t, alice.httpExpect, alice.id.String(), txID, "success")

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

	if documentType == typePO {
		documentType = poPrefix
	}

	objProof.Path("$.header.document_id").String().Equal(docIdentifier)
	objProof.Path("$.field_proofs[0].property").String().Equal(documentType + ".net_amount")
	objProof.Path("$.field_proofs[0].sorted_hashes").NotNull()
	objProof.Path("$.field_proofs[1].property").String().Equal(documentType + ".currency")
	objProof.Path("$.field_proofs[1].sorted_hashes").NotNull()

}
