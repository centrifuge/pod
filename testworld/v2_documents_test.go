// +build testworld

package testworld

import (
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	"github.com/centrifuge/go-centrifuge/http/coreapi"
	"github.com/centrifuge/go-centrifuge/testingutils"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/stretchr/testify/assert"
)

func TestV2GenericCreateAndCommit_new_document(t *testing.T) {
	t.Parallel()
	createNewDocument(t, func(dids []string) (map[string]interface{}, map[string]string) {
		return genericCoreAPICreate(dids), nil
	}, func(dids []string) (map[string]interface{}, map[string]string) {
		return genericCoreAPIUpdate(dids), nil
	})
}

func TestV2GenericCloneDocument(t *testing.T) {
	t.Parallel()
	cloneNewDocument(t, func(dids []string) (map[string]interface{}, map[string]string) {
		return genericCoreAPICreate(dids), nil
	})
}

func TestV2GenericCreate_next_version(t *testing.T) {
	t.Parallel()
	createNextDocument(t, genericCoreAPICreate)
}

func TestV2EntityCreateAndCommit_new_document(t *testing.T) {
	t.Parallel()
	createNewDocument(t, func(dids []string) (map[string]interface{}, map[string]string) {
		params := map[string]string{
			"legal_name": "test company",
			"identity":   dids[0],
		}
		return entityCoreAPICreate(dids[0], dids), params
	}, func(dids []string) (map[string]interface{}, map[string]string) {
		p := entityCoreAPIUpdate(dids)
		params := map[string]string{
			"legal_name": "updated company",
		}

		return p, params
	})
}

func TestV2EntityCreate_next_version(t *testing.T) {
	t.Parallel()
	createNextDocument(t, func(dids []string) map[string]interface{} {
		var id string
		if len(dids) > 0 {
			id = dids[0]
		}
		return entityCoreAPICreate(id, dids)
	})
}

func createNewDocument(
	t *testing.T,
	createPayloadParams, updatePayloadParams func([]string) (map[string]interface{}, map[string]string)) {
	alice := doctorFord.getHostTestSuite(t, "Alice")
	bob := doctorFord.getHostTestSuite(t, "Bob")
	charlie := doctorFord.getHostTestSuite(t, "Charlie")

	// Alice prepares document to share with Bob and charlie
	payload, params := createPayloadParams([]string{bob.id.String(), charlie.id.String()})
	res := createDocument(alice.httpExpect, alice.id.String(), "documents", http.StatusCreated, payload)
	status := getDocumentStatus(t, res)
	assert.Equal(t, status, "pending")

	checkDocumentParams(res, params)
	label := "signed_attribute"
	signedAttributeMissing(t, res, label)
	docID := getDocumentIdentifier(t, res)
	assert.NotEmpty(t, docID)

	// getting pending document should be successful
	getV2DocumentWithStatus(alice.httpExpect, alice.id.String(), docID, "pending", http.StatusOK)

	// committed shouldn't be success
	getV2DocumentWithStatus(alice.httpExpect, alice.id.String(), docID, "committed", http.StatusNotFound)

	// add a signed attribute
	value := hexutil.Encode(utils.RandomSlice(32))
	res = addSignedAttribute(alice.httpExpect, alice.id.String(), docID, label, value, "bytes")
	signedAttributeExists(t, res, label)

	// Alice updates the document
	payload, params = updatePayloadParams([]string{bob.id.String(), charlie.id.String()})
	payload["document_id"] = docID
	res = updateDocument(alice.httpExpect, alice.id.String(), "documents", http.StatusOK, payload)
	status = getDocumentStatus(t, res)
	assert.Equal(t, status, "pending")
	checkDocumentParams(res, params)
	getV2DocumentWithStatus(alice.httpExpect, alice.id.String(), docID, "pending", http.StatusOK)

	// alice removes charlie from the list of collaborators
	removeCollaborators(alice.httpExpect, alice.id.String(), "documents", http.StatusOK, docID, charlie.id.String())

	// Commits document and shares with Bob
	res = commitDocument(alice.httpExpect, alice.id.String(), "documents", http.StatusAccepted, docID)
	jobID := getJobID(t, res)
	ok, err := waitForJobComplete(alice.httpExpect, alice.id.String(), jobID)
	assert.NoError(t, err)
	assert.True(t, ok)
	getDocumentAndVerify(t, alice.httpExpect, alice.id.String(), docID, nil, updateAttributes())

	// pending document should fail
	getV2DocumentWithStatus(alice.httpExpect, alice.id.String(), docID, "pending", http.StatusNotFound)

	// committed should be successful
	getV2DocumentWithStatus(alice.httpExpect, alice.id.String(), docID, "committed", http.StatusOK)

	// Bob should have the document
	getDocumentAndVerify(t, bob.httpExpect, bob.id.String(), docID, nil, updateAttributes())

	// charlie should not have the document
	nonExistingDocumentCheck(charlie.httpExpect, charlie.id.String(), docID)

	// try to commit same document again - failure
	commitDocument(alice.httpExpect, alice.id.String(), "documents", http.StatusBadRequest, docID)
}

func createNextDocument(t *testing.T, createPayload func([]string) map[string]interface{}) {
	alice := doctorFord.getHostTestSuite(t, "Alice")
	bob := doctorFord.getHostTestSuite(t, "Bob")

	// Alice shares document with Bob
	docID := createAndCommitDocument(t, alice.httpExpect, alice.id.String(), createPayload([]string{bob.id.String()}))
	res := getDocumentAndVerify(t, bob.httpExpect, bob.id.String(), docID, nil, createAttributes()).Object()
	versionID := getDocumentCurrentVersion(t, res)
	assert.Equal(t, docID, versionID, "failed to create a fresh document")

	// there should be no pending document with alice
	getV2DocumentWithStatus(alice.httpExpect, alice.id.String(), docID, "pending", http.StatusNotFound)

	// bob creates a next pending version of the document
	payload := createPayload(nil)
	payload["document_id"] = docID
	res = createDocument(bob.httpExpect, bob.id.String(), "documents", http.StatusCreated, payload)
	status := getDocumentStatus(t, res)
	assert.Equal(t, status, "pending", "document must be in pending status")
	edocID := getDocumentIdentifier(t, res)
	assert.Equal(t, docID, edocID, "document identifiers mismatch")
	eversionID := getDocumentCurrentVersion(t, res)
	assert.NotEqual(t, docID, eversionID, "document ID and versionID must not be equal")
	// alice should not have this version
	nonExistingDocumentVersionCheck(alice.httpExpect, alice.id.String(), docID, eversionID)

	// bob has pending document
	getV2DocumentWithStatus(bob.httpExpect, bob.id.String(), docID, "pending", http.StatusOK)

	// commit the document
	// Commits document and shares with alice
	res = commitDocument(bob.httpExpect, bob.id.String(), "documents", http.StatusAccepted, docID)
	jobID := getJobID(t, res)
	ok, err := waitForJobComplete(bob.httpExpect, bob.id.String(), jobID)
	assert.NoError(t, err)
	assert.True(t, ok)

	// bob shouldn't have any pending documents but has a committed one
	getV2DocumentWithStatus(bob.httpExpect, bob.id.String(), docID, "pending", http.StatusNotFound)
	getV2DocumentWithStatus(bob.httpExpect, bob.id.String(), docID, "committed", http.StatusOK)
	getV2DocumentWithStatus(alice.httpExpect, alice.id.String(), docID, "committed", http.StatusOK)
}

func cloneNewDocument(
	t *testing.T,
	createPayloadParams func([]string) (map[string]interface{}, map[string]string)) {
	alice := doctorFord.getHostTestSuite(t, "Alice")
	bob := doctorFord.getHostTestSuite(t, "Bob")

	// Alice prepares document to share with Bob
	payload, _ := createPayloadParams([]string{bob.id.String()})
	res := createDocument(alice.httpExpect, alice.id.String(), "documents", http.StatusCreated, payload)
	status := getDocumentStatus(t, res)
	assert.Equal(t, status, "pending")

	docID := getDocumentIdentifier(t, res)
	assert.NotEmpty(t, docID)

	// getting pending document should be successful
	getV2DocumentWithStatus(alice.httpExpect, alice.id.String(), docID, "pending", http.StatusOK)

	// Commits template
	res = commitDocument(alice.httpExpect, alice.id.String(), "documents", http.StatusAccepted, docID)
	jobID := getJobID(t, res)
	ok, err := waitForJobComplete(alice.httpExpect, alice.id.String(), jobID)
	assert.NoError(t, err)
	assert.True(t, ok)
	getDocumentAndVerify(t, alice.httpExpect, alice.id.String(), docID, nil, createAttributes())

	// Bob should have the template
	getDocumentAndVerify(t, bob.httpExpect, bob.id.String(), docID, nil, createAttributes())

	// Bob clones the document from a payload with a template ID
	valid := map[string]interface{}{
		"scheme":      "generic",
		"document_id": docID,
	}

	res1 := cloneDocument(bob.httpExpect, bob.id.String(), "documents", http.StatusCreated, valid)
	docID1 := getDocumentIdentifier(t, res1)
	assert.NotEmpty(t, docID1)
	res = commitDocument(bob.httpExpect, bob.id.String(), "documents", http.StatusAccepted, docID1)
	jobID = getJobID(t, res)
	ok, err = waitForJobComplete(bob.httpExpect, bob.id.String(), jobID)
	assert.NoError(t, err)
	assert.True(t, ok)

	getClonedDocumentAndCheck(t, bob.httpExpect, bob.id.String(), docID, docID1, nil, createAttributes())
}

func TestDocument_ComputeFields(t *testing.T) {
	alice := doctorFord.getHostTestSuite(t, "Alice")
	bob := doctorFord.getHostTestSuite(t, "Bob")

	payload := genericCoreAPICreate([]string{alice.id.String()})
	res := createDocument(alice.httpExpect, alice.id.String(), "documents", http.StatusCreated, payload)
	status := getDocumentStatus(t, res)
	assert.Equal(t, status, "pending")

	docID := getDocumentIdentifier(t, res)
	assert.NotEmpty(t, docID)

	wasm, err := ioutil.ReadFile("../testingutils/compute_fields/simple_average.wasm")
	assert.NoError(t, err)

	// create role
	roleID := utils.RandomSlice(32)
	obj := addRole(alice.httpExpect, alice.id.String(), docID, hexutil.Encode(roleID), []string{bob.id.String()}, http.StatusOK)
	r, cs := parseRole(obj)
	assert.Equal(t, r, hexutil.Encode(roleID))
	assert.Contains(t, cs, strings.ToLower(bob.id.String()))

	// set compute fields
	rules := map[string][]map[string]interface{}{
		"compute_fields_rules": {
			{
				"wasm":                   hexutil.Encode(wasm),
				"attribute_labels":       []string{"test", "test1", "test2"},
				"target_attribute_label": "result",
			},
		},
		"attribute_rules": {
			{
				"key_label": "test",
				"role_id":   hexutil.Encode(roleID),
			},
			{
				"key_label": "test1",
				"role_id":   hexutil.Encode(roleID),
			},
			{
				"key_label": "test2",
				"role_id":   hexutil.Encode(roleID),
			},
		},
	}
	obj = addTransitionRules(alice.httpExpect, alice.id.String(), docID, rules, http.StatusOK)
	tr := parseRules(t, obj)
	assert.Len(t, tr.Rules, 4)
	ruleID := tr.Rules[3].RuleID[:]
	obj = getTransitionRule(alice.httpExpect, alice.id.String(), docID, hexutil.Encode(ruleID), http.StatusOK)
	rule := parseRule(t, obj)
	assert.Equal(t, tr.Rules[3], rule)

	// commits the document
	res = commitDocument(alice.httpExpect, alice.id.String(), "documents", http.StatusAccepted, docID)
	jobID := getJobID(t, res)
	ok, err := waitForJobComplete(alice.httpExpect, alice.id.String(), jobID)
	assert.NoError(t, err)
	assert.True(t, ok)
	var result [32]byte
	getDocumentAndVerify(t, alice.httpExpect, alice.id.String(), docID, nil, withComputeFieldResultAttribute(result[:]))
	getDocumentAndVerify(t, bob.httpExpect, bob.id.String(), docID, nil, withComputeFieldResultAttribute(result[:]))

	// bob adds the attributes
	p := genericCoreAPICreate(nil)
	attrs := coreapi.AttributeMapRequest{
		"test": coreapi.AttributeRequest{
			Type:  "integer",
			Value: "1000",
		},
		"test1": coreapi.AttributeRequest{
			Type:  "integer",
			Value: "2000",
		},
		"test2": coreapi.AttributeRequest{
			Type:  "integer",
			Value: "3000",
		},
	}
	p["attributes"] = attrs
	p["document_id"] = docID
	createAndCommitDocument(t, bob.httpExpect, bob.id.String(), p)

	// result = encoded(risk(1)) + encoded((1000+2000+3000)/3 = 1000)
	result = [32]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0x7, 0xd0}
	reqAttrs := withComputeFieldResultAttribute(result[:])
	for k, v := range attrs {
		reqAttrs[k] = v
	}

	getDocumentAndVerify(t, alice.httpExpect, alice.id.String(), docID, nil, reqAttrs)
	getDocumentAndVerify(t, bob.httpExpect, bob.id.String(), docID, nil, reqAttrs)
}

func TestPushToOracle(t *testing.T) {
	docID, tokenID := defaultNFTMint(t)
	alice := doctorFord.getHostTestSuite(t, "Alice")
	fp := getFingerprint(t, alice.httpExpect, alice.id.String(), docID)
	oracle, err := testingutils.DeployOracleContract(fp, alice.id.String())
	assert.NoError(t, err)
	payload := map[string]string{
		"token_id":       tokenID.String(),
		"attribute_key":  "0xf6a214f7a5fcda0c2cee9660b7fc29f5649e3c68aad48e20e950137c98913a68",
		"oracle_address": oracle,
	}
	obj := pushToOracle(alice.httpExpect, alice.id.String(), docID, payload, http.StatusAccepted)
	jobID := obj.Raw()["job_id"].(string)
	ok, err := waitForJobComplete(alice.httpExpect, alice.id.String(), jobID)
	assert.NoError(t, err)
	assert.True(t, ok)
}
