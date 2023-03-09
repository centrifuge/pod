//go:build testworld

package testworld

import (
	"github.com/centrifuge/pod/documents"
	"github.com/centrifuge/pod/http/coreapi"
	"github.com/centrifuge/pod/utils"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

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
	if documentType != "documents" {
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

func entityCoreAPICreate(identity string, collaborators []string) map[string]interface{} {
	p := map[string]interface{}{
		"scheme":       "entity",
		"write_access": collaborators,
		"attributes":   createAttributes(),
	}

	data := map[string]interface{}{
		"legal_name": "test company",
		"contacts": []map[string]interface{}{
			{
				"name": "test name",
			},
		},

		"payment_details": []map[string]interface{}{
			{
				"predefined": true,
				"bank_payment_method": map[string]interface{}{
					"identifier":  hexutil.Encode(utils.RandomSlice(32)),
					"holder_name": "John Doe",
				},
			},
		},
	}

	if identity != "" {
		data["identity"] = identity
	}

	p["data"] = data
	return p
}

func entityCoreAPIUpdate(collabs []string) map[string]interface{} {
	p := map[string]interface{}{
		"scheme":       "entity",
		"write_access": collabs,
		"data": map[string]interface{}{
			"legal_name": "updated company",
		},
		"attributes": updateAttributes(),
	}

	return p
}

func genericCoreAPICreate(collaborators []string) map[string]interface{} {
	return map[string]interface{}{
		"scheme":       "generic",
		"write_access": collaborators,
		"data":         map[string]interface{}{},
		"attributes":   createAttributes(),
	}
}

func genericCoreAPIUpdate(collaborators []string) map[string]interface{} {
	payload := genericCoreAPICreate(collaborators)
	payload["attributes"] = updateAttributes()
	return payload
}

func createAttributes() coreapi.AttributeMapRequest {
	dec, _ := documents.NewDecimal("100001.002")
	return coreapi.AttributeMapRequest{
		"string_test": coreapi.AttributeRequest{
			Type:  "string",
			Value: "hello, world",
		},
		"monetary_test": coreapi.AttributeRequest{
			Type: "monetary",
			MonetaryValue: &coreapi.MonetaryValue{
				Value: dec,
				ID:    "USD",
			},
		},
	}
}

func withComputeFieldResultAttribute(res []byte) coreapi.AttributeMapRequest {
	dec, _ := documents.NewDecimal("100001.002")
	return coreapi.AttributeMapRequest{
		"string_test": coreapi.AttributeRequest{
			Type:  "string",
			Value: "hello, world",
		},
		"monetary_test": coreapi.AttributeRequest{
			Type: "monetary",
			MonetaryValue: &coreapi.MonetaryValue{
				Value: dec,
				ID:    "USD",
			},
		},
		"result": coreapi.AttributeRequest{
			Type:  "bytes",
			Value: hexutil.Encode(res),
		},
	}
}

func updateAttributes() coreapi.AttributeMapRequest {
	return coreapi.AttributeMapRequest{
		"decimal_test": coreapi.AttributeRequest{
			Type:  "decimal",
			Value: "100.001",
		},
	}
}

func allAttributes() coreapi.AttributeMapRequest {
	attrs := createAttributes()
	for k, v := range updateAttributes() {
		attrs[k] = v
	}

	return attrs
}
