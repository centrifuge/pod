// +build unit

package documents

//func TestCreateProofs(t *testing.T) {
//	CreateProofs()
//	i, corDoc, err := createMockInvoice(t)
//	assert.Nil(t, err)
//	corDoc, proof, err := i.createProofs([]string{"invoice_number", "collaborators[0]", "document_type"})
//	assert.Nil(t, err)
//	assert.NotNil(t, proof)
//	assert.NotNil(t, corDoc)
//	tree, _ := coredocument.GetDocumentRootTree(corDoc)
//
//	// Validate invoice_number
//	valid, err := tree.ValidateProof(proof[0])
//	assert.Nil(t, err)
//	assert.True(t, valid)
//
//	// Validate collaborators[0]
//	valid, err = tree.ValidateProof(proof[1])
//	assert.Nil(t, err)
//	assert.True(t, valid)
//
//	// Validate document_type
//	valid, err = tree.ValidateProof(proof[2])
//	assert.Nil(t, err)
//	assert.True(t, valid)
//}
//
//func createMockCoreDocument(t *testing.T) *coredocumentpb.CoreDocument {
//
//	return
//}
