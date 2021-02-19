// +build testworld

package testworld

import (
	"testing"

	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/http/coreapi"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

func TestCoreAPI_DocumentGenericCreateAndUpdate(t *testing.T) {
	t.Parallel()
	alice := doctorFord.getHostTestSuite(t, "Alice")
	bob := doctorFord.getHostTestSuite(t, "Bob")
	charlie := doctorFord.getHostTestSuite(t, "Charlie")

	// Alice shares document with Bob first
	docID := createAndCommitDocument(t, doctorFord.maeve, alice.httpExpect, alice.id.String(),
		genericCoreAPICreate([]string{bob.id.String()}))
	params := map[string]interface{}{}
	getDocumentAndVerify(t, alice.httpExpect, alice.id.String(), docID, params, createAttributes())
	getDocumentAndVerify(t, bob.httpExpect, bob.id.String(), docID, params, createAttributes())
	nonExistingDocumentCheck(charlie.httpExpect, charlie.id.String(), docID)

	// Bob updates purchase order and shares with Charlie as well
	payload := genericCoreAPIUpdate([]string{alice.id.String(), charlie.id.String()})
	payload["document_id"] = docID
	docID = createAndCommitDocument(t, doctorFord.maeve, bob.httpExpect, bob.id.String(), payload)
	getDocumentAndVerify(t, alice.httpExpect, alice.id.String(), docID, params, allAttributes())
	getDocumentAndVerify(t, bob.httpExpect, bob.id.String(), docID, params, allAttributes())
	getDocumentAndVerify(t, charlie.httpExpect, charlie.id.String(), docID, params, allAttributes())
}

func TestCoreAPI_DocumentEntityCreateAndUpdate(t *testing.T) {
	t.Parallel()
	alice := doctorFord.getHostTestSuite(t, "Alice")
	bob := doctorFord.getHostTestSuite(t, "Bob")
	charlie := doctorFord.getHostTestSuite(t, "Charlie")

	// Alice shares document with Bob first
	docID := createAndCommitDocument(t, doctorFord.maeve, alice.httpExpect, alice.id.String(), entityCoreAPICreate(alice.id.String(), []string{bob.id.String(), charlie.id.String()}))
	params := map[string]interface{}{
		"identity":   alice.id.String(),
		"legal_name": "test company",
	}

	getDocumentAndVerify(t, alice.httpExpect, alice.id.String(), docID, params, createAttributes())
	getDocumentAndVerify(t, bob.httpExpect, bob.id.String(), docID, params, createAttributes())
	getDocumentAndVerify(t, charlie.httpExpect, charlie.id.String(), docID, params, createAttributes())
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
