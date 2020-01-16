// +build testworld

package testworld

func defaultEntityPayload(identity string, collaborators []string) map[string]interface{} {
	return map[string]interface{}{
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

func defaultRelationshipPayload(identity, targetID string) map[string]interface{} {
	return map[string]interface{}{
		"identity":        identity,
		"target_identity": targetID,
	}
}

func defaultFundingPayload(borrowerId, funderId string) map[string]interface{} {
	return map[string]interface{}{
		"data": map[string]interface{}{
			"amount":             "20000",
			"apr":                "0.33",
			"days":               "90",
			"currency":           "USD",
			"fee":                "30.30",
			"repayment_due_date": "2018-09-26T23:12:37.902198664Z",
			"borrower_id":        borrowerId,
			"funder_id":          funderId,
		},
	}
}

func defaultTransferPayload(senderId, recipientId string) map[string]interface{} {
	return map[string]interface{}{
		"data": map[string]interface{}{
			"status":         "open",
			"currency":       "EUR",
			"amount":         "300",
			"scheduled_date": "2018-09-26T23:12:37Z",
			"sender_id":      senderId,
			"recipient_id":   recipientId,
		},
	}
}

func updateTransferPayload(senderId, recipientId string) map[string]interface{} {
	return map[string]interface{}{
		"data": map[string]interface{}{
			"status":         "settled",
			"currency":       "EUR",
			"amount":         "400",
			"scheduled_date": "2018-09-26T23:12:37Z",
			"sender_id":      senderId,
			"recipient_id":   recipientId,
		},
	}
}

func updateFundingPayload(agreementId, borrowerId, funderId string) map[string]interface{} {
	return map[string]interface{}{
		"data": map[string]interface{}{
			"agreement_id":       agreementId,
			"funder_id":          funderId,
			"borrower_id":        borrowerId,
			"amount":             "10000",
			"apr":                "0.55",
			"days":               "90",
			"currency":           "USD",
			"fee":                "30.30",
			"repayment_due_date": "2018-09-26T23:12:37.902198664Z",
		},
	}
}

func updatedEntityPayload(identity string, collaborators []string) map[string]interface{} {
	return map[string]interface{}{
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
