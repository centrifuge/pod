//go:build testworld
// +build testworld

package testworld

//func TestProofWithMultipleFields_invoice_successful(t *testing.T) {
//	t.Parallel()
//	proofWithMultipleFieldsSuccessful(t, typeDocuments)
//}
//
//func proofWithMultipleFieldsSuccessful(t *testing.T, documentType string) {
//	alice := doctorFord.getHostTestSuite(t, "Alice")
//	bob := doctorFord.getHostTestSuite(t, "Bob")
//
//	// Alice shares a document with Bob
//	docID := createAndCommitDocument(t, doctorFord.maeve, alice.httpExpect, alice.id.String(), genericCoreAPICreate([]string{bob.id.String()}))
//	proofPayload := defaultProofPayload(documentType)
//	proofFromAlice := getProof(alice.httpExpect, alice.id.String(), http.StatusOK, docID, proofPayload)
//	proofFromBob := getProof(bob.httpExpect, bob.id.String(), http.StatusOK, docID, proofPayload)
//
//	checkProof(proofFromAlice, docID)
//	checkProof(proofFromBob, docID)
//}
//
//func checkProof(objProof *httpexpect.Object, docIdentifier string) {
//	prop1 := "0x0005000000000001" // generic.Scheme
//	prop2 := "0x0100000000000009" // cd_tree.documentIdentifier
//	objProof.Path("$.header.document_id").String().Equal(docIdentifier)
//	objProof.Path("$.field_proofs[0].property").String().Equal(prop1)
//	objProof.Path("$.field_proofs[0].sorted_hashes").NotNull()
//	objProof.Path("$.field_proofs[1].property").String().Equal(prop2)
//	objProof.Path("$.field_proofs[1].sorted_hashes").NotNull()
//}
