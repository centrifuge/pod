// +build testworld

package testworld

import (
	"net/http"
	"testing"
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

	// Alice want's to get a proof
	proofPayload := map[string]interface{}{
		"type":   "http://github.com/centrifuge/centrifuge-protobufs/invoice/#invoice.InvoiceData",
		"fields": []string{"invoice.net_amount", "invoice.currency"},
	}
	objProof := getProof(alice.httpExpect, http.StatusOK, docIdentifier, proofPayload)

	objProof.Path("$.header.document_id").String().Equal(docIdentifier)
	objProof.Path("$.field_proofs[0].property").String().Equal("invoice.net_amount")
	objProof.Path("$.field_proofs[0].sorted_hashes").NotNull()
	objProof.Path("$.field_proofs[1].property").String().Equal("invoice.currency")
	objProof.Path("$.field_proofs[1].sorted_hashes").NotNull()

}
