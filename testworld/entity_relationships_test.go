//go:build testworld

package testworld

//func TestHost_Entity_EntityRelationships(t *testing.T) {
//	t.Parallel()
//
//	// Hosts
//	alice := doctorFord.getHostTestSuite(t, "Alice")
//	bob := doctorFord.getHostTestSuite(t, "Bob")
//	charlie := doctorFord.getHostTestSuite(t, "Charlie")
//
//	// Alice anchors entity
//	docID := createAndCommitDocument(t, doctorFord.maeve, alice.httpExpect, alice.id.ToHexString(), defaultEntityPayload(alice.id.ToHexString(), []string{}))
//
//	// Alice creates an EntityRelationship with Bob
//	erID := createAndCommitDocument(t, doctorFord.maeve, alice.httpExpect, alice.id.ToHexString(),
//		defaultRelationshipPayload(alice.id.ToHexString(), docID,
//			bob.id.ToHexString()))
//
//	// Charlie should not have access to the entity data
//	nonExistingDocumentCheck(charlie.httpExpect, charlie.id.ToHexString(), erID)
//
//	// Bob should have access to the Entity through the EntityRelationship
//	response := getEntityWithRelation(bob.httpExpect, bob.id.ToHexString(), erID)
//	response.Path("$.data.legal_name").String().Equal("test company")
//
//	// Alice updates her entity
//	payload := updatedEntityPayload(alice.id.ToHexString(), []string{})
//	payload["document_id"] = docID
//	docID = createAndCommitDocument(t, doctorFord.maeve, alice.httpExpect, alice.id.ToHexString(), payload)
//
//	// Bob accesses the Entity through the EntityRelationship with Alice, this should return him the latest/updated Entity data
//	response = getEntityWithRelation(bob.httpExpect, bob.id.ToHexString(), erID)
//	response.Path("$.data.legal_name").String().Equal("edited test company")
//
//	// Alice wants to list all relationships associated with her entity, this should return her one (with Bob)
//	response = getEntityRelationships(alice.httpExpect, alice.id.ToHexString(), docID)
//	relationships := response.Array()
//	assert.Len(t, relationships.Raw(), 1)
//	relationship := relationships.Element(0)
//	relationship.Path("$.data.target_identity").String().Equal(bob.id.ToHexString())
//
//	// Alice creates an EntityRelationship with Charlie
//	cerID := createAndCommitDocument(t, doctorFord.maeve, alice.httpExpect, alice.id.ToHexString(),
//		defaultRelationshipPayload(alice.id.ToHexString(), docID, charlie.id.ToHexString()))
//
//	// Charlie should now have access to the Entity Data
//	response = getEntityWithRelation(charlie.httpExpect, charlie.id.ToHexString(), cerID)
//	response.Path("$.data.legal_name").String().Equal("edited test company")
//
//	// Alice lists all relationship associated with her entity, this should return her two (with Bob and Charlie)
//	response = getEntityRelationships(alice.httpExpect, alice.id.ToHexString(), docID)
//	relationships = response.Array()
//	assert.Len(t, relationships.Raw(), 2)
//
//	// Alice revokes the EntityRelationship with Bob
//	payload = defaultRelationshipPayload(alice.id.ToHexString(), docID, bob.id.ToHexString())
//	payload["document_id"] = erID
//	erID = createAndCommitDocument(t, doctorFord.maeve, alice.httpExpect, alice.id.ToHexString(), payload)
//
//	// Bob should no longer have access to the EntityRelationship
//	nonexistentEntityWithRelation(bob.httpExpect, bob.id.ToHexString(), erID)
//
//	// Alice lists all relationships associated with her entity
//	// This should return her one relationship with charlie
//	response = getEntityRelationships(alice.httpExpect, alice.id.ToHexString(), docID)
//	relationships = response.Array()
//	assert.Len(t, relationships.Raw(), 1)
//}
