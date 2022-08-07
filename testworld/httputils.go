//go:build testworld

package testworld

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"testing"

	"github.com/centrifuge/go-centrifuge/http/coreapi"
	v2 "github.com/centrifuge/go-centrifuge/http/v2"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/gavv/httpexpect"
	"github.com/stretchr/testify/assert"
)

const typeDocuments string = "documents"

var isRunningOnCI = len(os.Getenv("TRAVIS")) != 0

type httpLog struct {
	logger httpexpect.Logger
}

func (h *httpLog) Logf(fm string, args ...interface{}) {
	if !isRunningOnCI {
		h.logger.Logf(fm, args...)
	}
}

func createInsecureClientWithExpect(t *testing.T, baseURL string) *httpexpect.Expect {
	config := httpexpect.Config{
		BaseURL:  baseURL,
		Client:   createInsecureClient(),
		Reporter: httpexpect.NewAssertReporter(t),
		Printers: []httpexpect.Printer{
			httpexpect.NewCompactPrinter(&httpLog{t}),
		},
	}
	return httpexpect.WithConfig(config)
}

func createAndCommitDocument(t *testing.T, maeve *webhookReceiver, e *httpexpect.Expect, auth string,
	payload map[string]interface{}) (docID string) {
	res := createDocument(e, auth, "documents", http.StatusCreated, payload)
	docID = getDocumentIdentifier(t, res)
	res = commitDocument(e, auth, "documents", http.StatusAccepted, docID)
	jobID := getJobID(t, res)
	err := waitForJobComplete(maeve, e, auth, jobID)
	assert.NoError(t, err)
	return docID
}

func getEntityRelationships(e *httpexpect.Expect, auth string, docIdentifier string) *httpexpect.Value {
	objGet := addCommonHeaders(e.GET("/v2/entities/"+docIdentifier+"/relationships"), auth).
		Expect().Status(http.StatusOK).JSON().NotNull()

	return objGet
}

func getEntityWithRelation(e *httpexpect.Expect, auth string, relationshipID string) *httpexpect.Value {
	objGet := addCommonHeaders(e.GET("/v2/relationships/"+relationshipID+"/entity"), auth).
		Expect().Status(http.StatusOK).JSON().NotNull()

	return objGet
}

func nonexistentEntityWithRelation(e *httpexpect.Expect, auth string, relationshipID string) *httpexpect.Value {
	objGet := addCommonHeaders(e.GET("/v2/relationships/"+relationshipID+"/entity"), auth).
		Expect().Status(http.StatusNotFound).JSON().NotNull()

	return objGet
}

func nonExistingDocumentCheck(e *httpexpect.Expect, auth string, docIdentifier string) *httpexpect.Value {
	objGet := addCommonHeaders(e.GET("/v2/documents/"+docIdentifier+"/committed"), auth).
		Expect().Status(http.StatusNotFound).JSON().NotNull()
	return objGet
}

func nonExistingDocumentVersionCheck(e *httpexpect.Expect, auth string, docID, versionID string) *httpexpect.Value {
	objGet := addCommonHeaders(e.GET("/v2/documents/"+docID+"/versions/"+versionID), auth).
		Expect().Status(http.StatusNotFound).JSON().NotNull()
	return objGet
}

func createDocument(e *httpexpect.Expect, auth string, documentType string, status int, payload map[string]interface{}) *httpexpect.Object {
	obj := addCommonHeaders(e.POST("/v2/"+documentType), auth).
		WithJSON(payload).
		Expect().Status(status).JSON().Object()
	return obj
}

func cloneDocument(e *httpexpect.Expect, auth string, documentType string, status int, payload map[string]interface{}) *httpexpect.Object {
	obj := addCommonHeaders(e.POST("/v2/"+documentType+"/"+payload["document_id"].(string)+"/clone"), auth).
		WithJSON(payload).
		Expect().Status(status).JSON().Object()
	return obj
}

func updateDocument(e *httpexpect.Expect, auth string, documentType string, status int, payload map[string]interface{}) *httpexpect.Object {
	obj := addCommonHeaders(e.PATCH("/v2/"+documentType+"/"+payload["document_id"].(string)), auth).
		WithJSON(payload).
		Expect().Status(status).JSON().Object()
	return obj
}

func removeCollaborators(e *httpexpect.Expect, auth string, docType string, status int, docID string, collabs ...string) *httpexpect.Object {
	obj := addCommonHeaders(e.DELETE("/v2/"+docType+"/"+docID+"/collaborators"), auth).
		WithJSON(map[string][]string{
			"collaborators": collabs,
		}).
		Expect().Status(status).JSON().Object()
	return obj
}

func checkDocumentParams(obj *httpexpect.Object, params map[string]string) {
	for k, v := range params {
		obj.Path("$.data." + k).String().Equal(v)
	}
}

func commitDocument(e *httpexpect.Expect, auth string, documentType string, status int, docIdentifier string) *httpexpect.Object {
	obj := addCommonHeaders(e.POST("/v2/"+documentType+"/"+docIdentifier+"/commit"), auth).
		Expect().Status(status).JSON().Object()
	return obj
}

func getDocumentIdentifier(t *testing.T, response *httpexpect.Object) string {
	docIdentifier := response.Value("header").Path("$.document_id").String().NotEmpty().Raw()
	return docIdentifier
}

func getDocumentStatus(t *testing.T, response *httpexpect.Object) string {
	status := response.Value("header").Path("$.status").String().NotEmpty().Raw()
	return status
}

func getJobID(t *testing.T, resp *httpexpect.Object) string {
	txID := resp.Value("header").Path("$.job_id").String().Raw()
	if txID == "" {
		t.Error("transaction ID empty")
	}

	return txID
}

func getDocumentCurrentVersion(t *testing.T, resp *httpexpect.Object) string {
	versionID := resp.Value("header").Path("$.version_id").String().Raw()
	if versionID == "" {
		t.Error("version ID empty")
	}

	return versionID
}

func mintNFTV3(e *httpexpect.Expect, auth string, httpStatus int, payload map[string]interface{}) *httpexpect.Object {
	path := fmt.Sprintf("/v3/nfts/classes/%d/mint", payload["class_id"])
	resp := addCommonHeaders(e.POST(path), auth).
		WithJSON(payload).
		Expect().Status(httpStatus)

	httpObj := resp.JSON().Object()
	return httpObj
}

func ownerOfNFTV3(e *httpexpect.Expect, auth string, httpStatus int, payload map[string]interface{}) *httpexpect.Object {
	path := fmt.Sprintf(
		"/v3/nfts/classes/%d/instances/%s/owner",
		payload["class_id"],
		payload["instance_id"],
	)

	resp := addCommonHeaders(e.GET(path), auth).
		Expect().Status(httpStatus)

	httpObj := resp.JSON().Object()
	return httpObj
}

func metadataOfNFTV3(e *httpexpect.Expect, auth string, httpStatus int, payload map[string]interface{}) *httpexpect.Object {
	path := fmt.Sprintf(
		"/v3/nfts/classes/%d/instances/%s/metadata",
		payload["class_id"],
		payload["instance_id"],
	)

	resp := addCommonHeaders(e.GET(path), auth).
		Expect().Status(httpStatus)

	httpObj := resp.JSON().Object()
	return httpObj
}

func createNFTClassV3(e *httpexpect.Expect, auth string, httpStatus int, payload map[string]interface{}) *httpexpect.Object {
	resp := addCommonHeaders(e.POST("/v3/nfts/classes"), auth).
		WithJSON(payload).
		Expect().Status(httpStatus)

	httpObj := resp.JSON().Object()
	return httpObj
}

func mintNFT(e *httpexpect.Expect, auth string, httpStatus int, payload map[string]interface{}) *httpexpect.Object {
	resp := addCommonHeaders(e.POST("/v2/nfts/registries/"+payload["registry_address"].(string)+"/mint"), auth).
		WithJSON(payload).
		Expect().Status(httpStatus)

	httpObj := resp.JSON().Object()
	return httpObj
}

func mintNFTOnCC(e *httpexpect.Expect, auth string, httpStatus int, payload map[string]interface{}) *httpexpect.Object {
	resp := addCommonHeaders(e.POST("/beta/nfts/registries/"+payload["registry_address"].(string)+"/mint"), auth).
		WithJSON(payload).
		Expect().Status(httpStatus)

	httpObj := resp.JSON().Object()
	return httpObj
}

func transferNFT(e *httpexpect.Expect, auth string, httpStatus int, payload map[string]interface{}) *httpexpect.Object {
	resp := addCommonHeaders(e.POST("/v2/nfts/registries/"+payload["registry_address"].(string)+"/tokens/"+payload["token_id"].(string)+"/transfer"), auth).
		WithJSON(payload).
		Expect().Status(httpStatus)

	httpObj := resp.JSON().Object()
	return httpObj
}

func transferNFTOnCC(e *httpexpect.Expect, auth string, httpStatus int, payload map[string]interface{}) *httpexpect.
	Object {
	resp := addCommonHeaders(e.POST("/beta/nfts/registries/"+payload["registry_address"].(string)+"/tokens/"+payload["token_id"].(string)+"/transfer"), auth).
		WithJSON(payload).
		Expect().Status(httpStatus)

	httpObj := resp.JSON().Object()
	return httpObj
}

func ownerOfNFT(e *httpexpect.Expect, auth string, httpStatus int, payload map[string]interface{}) *httpexpect.Value {
	objGet := addCommonHeaders(e.GET("/v2/nfts/registries/"+payload["registry_address"].(string)+"/tokens/"+payload["token_id"].(string)+"/owner"), auth).
		Expect().Status(httpStatus).JSON().NotNull()
	return objGet
}

func ownerOfNFTOnCC(e *httpexpect.Expect, auth string, httpStatus int, payload map[string]interface{}) *httpexpect.Value {
	objGet := addCommonHeaders(e.GET("/beta/nfts/registries/"+payload["registry_address"].(string)+"/tokens/"+payload["token_id"].(string)+"/owner"), auth).
		Expect().Status(httpStatus).JSON().NotNull()
	return objGet
}

func getProof(e *httpexpect.Expect, auth string, httpStatus int, documentID string, payload map[string]interface{}) *httpexpect.Object {
	resp := addCommonHeaders(e.POST("/v2/documents/"+documentID+"/proofs"), auth).
		WithJSON(payload).
		Expect().Status(httpStatus)
	return resp.JSON().Object()
}

func getAccount(e *httpexpect.Expect, auth string, httpStatus int, identifier string) *httpexpect.Object {
	resp := addCommonHeaders(e.GET("/v2/accounts/"+identifier), auth).
		Expect().Status(httpStatus)
	return resp.JSON().Object()
}

func getAllAccounts(e *httpexpect.Expect, auth string, httpStatus int) *httpexpect.Object {
	resp := addCommonHeaders(e.GET("/v2/accounts"), auth).
		Expect().Status(httpStatus)
	return resp.JSON().Object()
}

func generateAccount(
	maeve *webhookReceiver,
	e *httpexpect.Expect,
	auth string,
	httpStatus int,
	payload map[string]any,
) (id *types.AccountID, err error) {
	req := addCommonHeaders(e.POST("/v2/accounts/generate"), auth).WithJSON(payload)
	resp := req.Expect()
	obj := resp.Status(httpStatus).JSON().Object().Raw()
	auth = obj["identity"].(string)
	jobID := obj["job_id"].(string)
	err = waitForJobComplete(maeve, e, auth, jobID)
	if err != nil {
		return nil, err
	}

	return types.NewAccountIDFromHexString(auth)
}

func createInsecureClient() *http.Client {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	return &http.Client{Transport: tr}
}

func waitForJobComplete(maeve *webhookReceiver, e *httpexpect.Expect, auth string, jobID string) error {
	ch := make(chan bool)
	maeve.waitForJobCompletion(jobID, ch)
	fmt.Println("Waiting for job notification", jobID)
	<-ch
	resp := addCommonHeaders(e.GET("/v2/jobs/"+jobID), auth).Expect().Status(200).JSON().Object()
	task := resp.Value("tasks").Array().Last().Object()
	message := task.Value("error").String().Raw()
	if message != "" {
		return errors.New(message)
	}

	return nil
}

func addCommonHeaders(req *httpexpect.Request, auth string) *httpexpect.Request {
	return req.
		WithHeader("accept", "application/json").
		WithHeader("Content-Type", "application/json").
		WithHeader("authorization", auth)
}

func getAccounts(accounts *httpexpect.Array) map[string]string {
	accIDs := make(map[string]string)
	for i := 0; i < int(accounts.Length().Raw()); i++ {
		val := accounts.Element(i).Path("$.identity_id").String().NotEmpty().Raw()
		accIDs[val] = val
	}
	return accIDs
}

func getFingerprint(t *testing.T, e *httpexpect.Expect, auth string, documentID string) string {
	obj := getDocumentAndVerify(t, e, auth, documentID, nil, nil)
	return obj.Path("$.header.fingerprint").String().Raw()
}

func getDocumentAndVerify(t *testing.T, e *httpexpect.Expect, auth string, documentID string, params map[string]interface{}, attrs coreapi.AttributeMapRequest) *httpexpect.Value {
	objGet := addCommonHeaders(e.GET("/v2/documents/"+documentID+"/committed"), auth).
		Expect().Status(http.StatusOK).JSON().NotNull()
	objGet.Path("$.header.document_id").String().Equal(documentID)
	for k, v := range params {
		objGet.Path("$.data." + k).String().Equal(v.(string))
	}

	if len(attrs) > 0 {
		reqJSON, err := json.Marshal(attrs)
		if err != nil {
			assert.Fail(t, err.Error())
		}

		gattrs := objGet.Path("$.attributes").Object().Raw()
		// Since we want to perform an equals check on the request attributes and response attributes we need to marshal and
		// unmarshal twice over the object
		respJSON, err := json.Marshal(gattrs)
		if err != nil {
			assert.Fail(t, err.Error())
		}
		var cattrs coreapi.AttributeMapRequest
		err = json.Unmarshal(respJSON, &cattrs)
		if err != nil {
			assert.Fail(t, err.Error())
		}
		respJSON, err = json.Marshal(cattrs)
		if err != nil {
			assert.Fail(t, err.Error())
		}

		assert.Equal(t, reqJSON, respJSON)
	}
	return objGet
}

func getClonedDocumentAndCheck(t *testing.T, e *httpexpect.Expect, auth string, docID string, docID1 string, params map[string]interface{}, attrs coreapi.AttributeMapRequest) *httpexpect.Value {
	objGet := addCommonHeaders(e.GET("/v2/documents/"+docID+"/committed"), auth).
		Expect().Status(http.StatusOK).JSON().NotNull()
	objGet.Path("$.header.document_id").String().Equal(docID)
	for k, v := range params {
		objGet.Path("$.data." + k).String().Equal(v.(string))
	}

	objGet1 := addCommonHeaders(e.GET("/v2/documents/"+docID1+"/committed"), auth).
		Expect().Status(http.StatusOK).JSON().NotNull()
	objGet1.Path("$.header.document_id").String().Equal(docID1)
	for k, v := range params {
		objGet.Path("$.data." + k).String().Equal(v.(string))
	}

	// make sure the fingerprints of the two documents are the same
	assert.Equal(t, objGet.Path("$.header.fingerprint"), objGet1.Path("$.header.fingerprint"))

	if len(attrs) > 0 {
		reqJson, err := json.Marshal(attrs)
		if err != nil {
			assert.Fail(t, err.Error())
		}

		gattrs := objGet.Path("$.attributes").Object().Raw()
		// Since we want to perform an equals check on the request attributes and response attributes we need to marshal and
		// unmarshal twice over the object
		respJson, err := json.Marshal(gattrs)
		if err != nil {
			assert.Fail(t, err.Error())
		}
		var cattrs coreapi.AttributeMapRequest
		err = json.Unmarshal(respJson, &cattrs)
		if err != nil {
			assert.Fail(t, err.Error())
		}
		respJson, err = json.Marshal(cattrs)
		if err != nil {
			assert.Fail(t, err.Error())
		}

		assert.Equal(t, reqJson, respJson)

		gattrs1 := objGet1.Path("$.attributes").Object().Raw()
		// Since we want to perform an equals check on the request attributes and response attributes we need to marshal and
		// unmarshal twice over the object
		respJson1, err := json.Marshal(gattrs1)
		if err != nil {
			assert.Fail(t, err.Error())
		}
		var cattrs1 coreapi.AttributeMapRequest
		err = json.Unmarshal(respJson1, &cattrs1)
		if err != nil {
			assert.Fail(t, err.Error())
		}
		respJson1, err = json.Marshal(cattrs1)
		if err != nil {
			assert.Fail(t, err.Error())
		}
		assert.Equal(t, gattrs, gattrs1)
	}
	return objGet
}

func nonExistingGenericDocumentVersionCheck(e *httpexpect.Expect, auth string, documentID, versionID string) *httpexpect.Value {
	objGet := addCommonHeaders(e.GET("/v2/documents/"+documentID+"/versions/"+versionID), auth).
		Expect().Status(404).JSON().NotNull()
	return objGet
}

func getV2DocumentWithStatus(e *httpexpect.Expect, auth, docID, status string, code int) *httpexpect.Value {
	objGet := addCommonHeaders(e.GET("/v2/documents/"+docID+"/"+status), auth).
		Expect().Status(code).JSON().NotNull()
	return objGet
}

func addSignedAttribute(e *httpexpect.Expect, auth, docID, label, payload, valType string) *httpexpect.Object {
	objPost := addCommonHeaders(e.POST("/v2/documents/"+docID+"/signed_attribute"), auth).WithJSON(map[string]string{
		"label":   label,
		"type":    valType,
		"payload": payload,
	}).Expect().Status(http.StatusOK).JSON().Object()
	return objPost
}

func signedAttributeExists(t *testing.T, res *httpexpect.Object, label string) {
	attrs := res.Path("$.attributes").Object().Raw()
	_, ok := attrs[label]
	assert.True(t, ok)
}

func signedAttributeMissing(t *testing.T, res *httpexpect.Object, label string) {
	attrs := res.Path("$.attributes").Object().Raw()
	_, ok := attrs[label]
	assert.False(t, ok)
}

func parseRole(obj *httpexpect.Object) (roleID string, collaborators []string) {
	roleID = obj.Path("$.id").String().Raw()
	for _, c := range obj.Path("$.collaborators").Array().Iter() {
		collaborators = append(collaborators, c.String().Raw())
	}
	return roleID, collaborators
}

func getRole(e *httpexpect.Expect, auth, docID, roleID string, status int) *httpexpect.Object {
	objPost := addCommonHeaders(e.GET("/v2/documents/"+docID+"/roles/"+roleID), auth).
		Expect().Status(status).JSON().Object()
	return objPost
}

func addRole(e *httpexpect.Expect, auth, docID, roleID string, collaborators []string, status int) *httpexpect.Object {
	objPost := addCommonHeaders(e.POST("/v2/documents/"+docID+"/roles"), auth).WithJSON(map[string]interface{}{
		"key":           roleID,
		"collaborators": collaborators,
	}).Expect().Status(status).JSON().Object()
	return objPost
}

func updateRole(e *httpexpect.Expect, auth, docID, roleID string, collaborators []string, status int) *httpexpect.Object {
	objPost := addCommonHeaders(e.PATCH("/v2/documents/"+docID+"/roles/"+roleID), auth).WithJSON(map[string]interface{}{
		"collaborators": collaborators,
	}).Expect().Status(status).JSON().Object()
	return objPost
}

func addTransitionRules(e *httpexpect.Expect, auth, docID string, payload map[string][]map[string]interface{}, status int) *httpexpect.Object {
	objPost := addCommonHeaders(e.POST("/v2/documents/"+docID+"/transition_rules"), auth).WithJSON(
		payload).Expect().Status(status).JSON().Object()
	return objPost
}

func getTransitionRule(e *httpexpect.Expect, auth, docID, ruleID string, status int) *httpexpect.Object {
	objPost := addCommonHeaders(e.GET("/v2/documents/"+docID+"/transition_rules/"+ruleID), auth).
		Expect().Status(status).JSON().Object()
	return objPost
}

func deleteTransitionRule(e *httpexpect.Expect, auth, docID, ruleID string, status int) *httpexpect.Response {
	objPost := addCommonHeaders(e.DELETE("/v2/documents/"+docID+"/transition_rules/"+ruleID), auth).
		Expect().Status(status)
	return objPost
}

func pushToOracle(e *httpexpect.Expect, auth, docID string, payload map[string]string, status int) *httpexpect.Object {
	resp := addCommonHeaders(e.POST("/v2/documents/"+docID+"/push_to_oracle"), auth).
		WithJSON(payload).
		Expect().Status(status)
	return resp.JSON().Object()
}

func parseRules(t *testing.T, obj *httpexpect.Object) v2.TransitionRules {
	d, err := json.Marshal(obj.Raw())
	assert.NoError(t, err)
	var tr v2.TransitionRules
	assert.NoError(t, json.Unmarshal(d, &tr))
	return tr
}

func parseRule(t *testing.T, obj *httpexpect.Object) v2.TransitionRule {
	d, err := json.Marshal(obj.Raw())
	assert.NoError(t, err)
	var tr v2.TransitionRule
	assert.NoError(t, json.Unmarshal(d, &tr))
	return tr
}
