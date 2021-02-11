// +build testworld

package testworld

func defaultEntityPayload(identity string, collaborators []string) map[string]interface{} {
	return map[string]interface{}{
		"scheme": "entity",
		"data": map[string]interface{}{
			"identity":   identity,
			"legal_name": "test company",
			"contacts": []map[string]interface{}{
				{
					"name": "test name",
				},
			},
		},
		"write_access": collaborators,
	}
}

func defaultRelationshipPayload(ownerDID, entityID, targetDID string) map[string]interface{} {
	return map[string]interface{}{
		"scheme": "entity_relationship",
		"data": map[string]interface{}{
			"owner_identity":    ownerDID,
			"entity_identifier": entityID,
			"target_identity":   targetDID,
		},
	}
}

func updatedEntityPayload(identity string, collaborators []string) map[string]interface{} {
	return map[string]interface{}{
		"scheme": "entity",
		"data": map[string]interface{}{
			"identity":   identity,
			"legal_name": "edited test company",
			"contacts": []map[string]interface{}{
				{
					"name": "test name",
				},
			},
		},
		"write_access": collaborators,
	}
}

func defaultProofPayload(documentType string) map[string]interface{} {
	if documentType != typeDocuments {
		return nil
	}

	return map[string]interface{}{
		"type":   "http://github.com/centrifuge/centrifuge-protobufs/generic/#generic.Generic",
		"fields": []string{"generic.scheme", "cd_tree.document_identifier"},
	}
}

func wrongGenericDocumentPayload(collabs []string) map[string]interface{} {
	return map[string]interface{}{
		"scheme":       "generic",
		"write_access": collabs,
		"data":         map[string]interface{}{},
		"attributes":   wrongAttributePayload(),
	}
}
func wrongAttributePayload() map[string]map[string]string {
	payload := map[string]map[string]string{
		"test_invalid": {
			"type":  "timestamp",
			"value": "some invalid time stamp",
		},
	}

	return payload
}
