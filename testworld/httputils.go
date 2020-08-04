// +build testworld

package testworld

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/centrifuge/go-centrifuge/httpapi/coreapi"
	v2 "github.com/centrifuge/go-centrifuge/httpapi/v2"
	"github.com/gavv/httpexpect"
	"github.com/stretchr/testify/assert"
)

const typeDocuments string = "documents"
const typeEntity string = "entities"

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
			httpexpect.NewCurlPrinter(&httpLog{t}),
		},
	}
	return httpexpect.WithConfig(config)
}

func getFundingAndCheck(e *httpexpect.Expect, auth, identifier, fundingID string, params map[string]interface{}) *httpexpect.Value {
	objGet := addCommonHeaders(e.GET("/v1/documents/"+identifier+"/funding_agreements/"+fundingID), auth).
		Expect().Status(http.StatusOK).JSON().NotNull()
	objGet.Path("$.header.document_id").String().Equal(identifier)
	objGet.Path("$.data.funding.currency").String().Equal(params["currency"].(string))
	objGet.Path("$.data.funding.amount").String().Equal(params["amount"].(string))
	objGet.Path("$.data.funding.apr").String().Equal(params["apr"].(string))
	return objGet
}

func getTransferAndCheck(e *httpexpect.Expect, auth, identifier, transferID string, params map[string]interface{}) *httpexpect.Value {
	objGet := addCommonHeaders(e.GET("/v1/documents/"+identifier+"/transfer_details/"+transferID), auth).
		Expect().Status(http.StatusOK).JSON().NotNull()
	objGet.Path("$.header.document_id").String().Equal(identifier)
	objGet.Path("$.data.status").String().Equal(params["status"].(string))
	objGet.Path("$.data.amount").String().Equal(params["amount"].(string))
	objGet.Path("$.data.scheduled_date").String().Equal(params["scheduled_date"].(string))
	return objGet
}

func getListTransfersCheck(e *httpexpect.Expect, auth, identifier string, listLen int, params map[string]interface{}) *httpexpect.Value {
	objGet := addCommonHeaders(e.GET("/v1/documents/"+identifier+"/transfer_details"), auth).
		Expect().Status(http.StatusOK).JSON().NotNull()
	objGet.Path("$.header.document_id").String().Equal(identifier)
	objGet.Path("$.data").Array().Length().Equal(listLen)

	for i := 0; i < listLen; i++ {
		objGet.Path("$.data").Array().Element(i).Path("$.status").String().Equal(params["status"].(string))
		objGet.Path("$.data").Array().Element(i).Path("$.amount").String().Equal(params["amount"].(string))
		objGet.Path("$.data").Array().Element(i).Path("$.scheduled_date").String().Equal(params["scheduled_date"].(string))
	}
	return objGet
}

func getListFundingCheck(e *httpexpect.Expect, auth, identifier string, listLen int, params map[string]interface{}) *httpexpect.Value {
	objGet := addCommonHeaders(e.GET("/v1/documents/"+identifier+"/funding_agreements"), auth).
		Expect().Status(http.StatusOK).JSON().NotNull()

	objGet.Path("$.header.document_id").String().Equal(identifier)
	objGet.Path("$.data").Array().Length().Equal(listLen)

	for i := 0; i < listLen; i++ {
		objGet.Path("$.data").Array().Element(i).Path("$.funding.currency").String().Equal(params["currency"].(string))
		objGet.Path("$.data").Array().Element(i).Path("$.funding.amount").String().Equal(params["amount"].(string))
		objGet.Path("$.data").Array().Element(i).Path("$.funding.apr").String().Equal(params["apr"].(string))
	}

	return objGet
}

func getFundingWithSignatureAndCheck(e *httpexpect.Expect, auth, identifier, agreementID, valid, outDatedSignature string, params map[string]interface{}) *httpexpect.Value {
	objGet := addCommonHeaders(e.GET("/v1/documents/"+identifier+"/funding_agreements/"+agreementID), auth).
		Expect().Status(http.StatusOK).JSON().NotNull()

	objGet.Path("$.header.document_id").String().Equal(identifier)
	objGet.Path("$.data.funding.currency").String().Equal(params["currency"].(string))
	objGet.Path("$.data.funding.amount").String().Equal(params["amount"].(string))
	objGet.Path("$.data.funding.apr").String().Equal(params["apr"].(string))

	objGet.Path("$.data.signatures").Array().Element(0).Path("$.valid").Equal(valid)
	objGet.Path("$.data.signatures").Array().Element(0).Path("$.outdated_signature").Equal(outDatedSignature)
	return objGet
}

func getEntityAndCheck(e *httpexpect.Expect, auth string, documentType string, params map[string]interface{}) *httpexpect.Value {
	docIdentifier := params["document_id"].(string)

	objGet := addCommonHeaders(e.GET("/v1/"+documentType+"/"+docIdentifier), auth).
		Expect().Status(http.StatusOK).JSON().NotNull()
	objGet.Path("$.header.document_id").String().Equal(docIdentifier)
	objGet.Path("$.data.entity.legal_name").String().Equal(params["legal_name"].(string))

	return objGet
}

func getEntity(e *httpexpect.Expect, auth string, docIdentifier string) *httpexpect.Value {

	objGet := addCommonHeaders(e.GET("/v1/entities/"+docIdentifier), auth).
		Expect().Status(http.StatusOK).JSON().NotNull()

	return objGet
}

func getEntityWithRelation(e *httpexpect.Expect, auth string, documentType string, params map[string]interface{}) *httpexpect.Value {
	relationshipIdentifier := params["r_identifier"].(string)

	objGet := addCommonHeaders(e.GET("/v1/relationships/"+relationshipIdentifier+"/entity"), auth).
		Expect().Status(http.StatusOK).JSON().NotNull()

	return objGet
}

func nonexistentEntityWithRelation(e *httpexpect.Expect, auth string, documentType string, params map[string]interface{}) *httpexpect.Value {
	relationshipIdentifier := params["r_identifier"].(string)

	objGet := addCommonHeaders(e.GET("/v1/relationships/"+relationshipIdentifier+"/entity"), auth).
		Expect().Status(http.StatusNotFound).JSON().NotNull()

	return objGet
}

func revokeEntity(e *httpexpect.Expect, auth, entityID string, status int, payload map[string]interface{}) *httpexpect.Object {
	obj := addCommonHeaders(e.POST("/v1/entities/"+entityID+"/revoke"), auth).
		WithJSON(payload).
		Expect().Status(status).JSON().Object()
	return obj
}

func nonExistingDocumentCheck(e *httpexpect.Expect, auth string, docIdentifier string) *httpexpect.Value {
	objGet := addCommonHeaders(e.GET("/v1/documents/"+docIdentifier), auth).
		Expect().Status(http.StatusNotFound).JSON().NotNull()
	return objGet
}

func nonExistingDocumentVersionCheck(e *httpexpect.Expect, auth string, docID, versionID string) *httpexpect.Value {
	objGet := addCommonHeaders(e.GET("/v1/documents/"+docID+"/versions/"+versionID), auth).
		Expect().Status(http.StatusNotFound).JSON().NotNull()
	return objGet
}

func createDocument(e *httpexpect.Expect, auth string, documentType string, status int, payload map[string]interface{}) *httpexpect.Object {
	obj := addCommonHeaders(e.POST("/v1/"+documentType), auth).
		WithJSON(payload).
		Expect().Status(status).JSON().Object()
	return obj
}

func createDocumentV2(e *httpexpect.Expect, auth string, documentType string, status int, payload map[string]interface{}) *httpexpect.Object {
	obj := addCommonHeaders(e.POST("/v2/"+documentType), auth).
		WithJSON(payload).
		Expect().Status(status).JSON().Object()
	return obj
}

func cloneDocumentV2(e *httpexpect.Expect, auth string, documentType string, status int, payload map[string]interface{}) *httpexpect.Object {
	obj := addCommonHeaders(e.POST("/v2/"+documentType+"/"+payload["document_id"].(string)+"/clone"), auth).
		WithJSON(payload).
		Expect().Status(status).JSON().Object()
	return obj
}

func updateDocumentV2(e *httpexpect.Expect, auth string, documentType string, status int, payload map[string]interface{}) *httpexpect.Object {
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

func updateCoreAPIDocument(e *httpexpect.Expect, auth string, documentType string, docID string, status int, payload map[string]interface{}) *httpexpect.Object {
	obj := addCommonHeaders(e.PUT("/v1/"+documentType+"/"+docID), auth).
		WithJSON(payload).
		Expect().Status(status)
	return obj.JSON().Object()
}

func createFunding(e *httpexpect.Expect, auth string, identifier string, status int, payload map[string]interface{}) *httpexpect.Object {
	obj := addCommonHeaders(e.POST("/v1/documents/"+identifier+"/funding_agreements"), auth).
		WithJSON(payload).
		Expect().Status(status).JSON().Object()
	return obj
}

func updateFunding(e *httpexpect.Expect, auth string, agreementId string, status int, docIdentifier string, payload map[string]interface{}) *httpexpect.Object {
	obj := addCommonHeaders(e.PUT("/v1/documents/"+docIdentifier+"/funding_agreements/"+agreementId), auth).
		WithJSON(payload).
		Expect().Status(status).JSON().Object()
	return obj
}

func createTransfer(e *httpexpect.Expect, auth string, identifier string, status int, payload map[string]interface{}) *httpexpect.Object {
	obj := addCommonHeaders(e.POST("/v1/documents/"+identifier+"/transfer_details"), auth).
		WithJSON(payload).
		Expect().Status(status).JSON().Object()
	return obj
}

func updateTransfer(e *httpexpect.Expect, auth string, status int, docIdentifier, transferId string, payload map[string]interface{}) *httpexpect.Object {
	obj := addCommonHeaders(e.PUT("/v1/documents/"+docIdentifier+"/transfer_details/"+transferId), auth).
		WithJSON(payload).
		Expect().Status(status).JSON().Object()
	return obj
}

func signFunding(e *httpexpect.Expect, auth, identifier, agreementId string, status int) *httpexpect.Object {
	obj := addCommonHeaders(e.POST("/v1/documents/"+identifier+"/funding_agreements/"+agreementId+"/sign"), auth).
		Expect().Status(status).JSON().Object()
	return obj
}

func shareEntity(e *httpexpect.Expect, auth, entityID string, status int, payload map[string]interface{}) *httpexpect.Object {
	obj := addCommonHeaders(e.POST("/v1/entities/"+entityID+"/share"), auth).
		WithJSON(payload).
		Expect().Status(status).JSON().Object()
	return obj
}

func updateDocument(e *httpexpect.Expect, auth string, documentType string, status int, docIdentifier string, payload map[string]interface{}) *httpexpect.Object {
	obj := addCommonHeaders(e.PUT("/v1/"+documentType+"/"+docIdentifier), auth).
		WithJSON(payload).
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

func getAgreementId(t *testing.T, response *httpexpect.Object) string {
	agreementID := response.Value("data").Path("$.funding.agreement_id").String().NotEmpty().Raw()
	if agreementID == "" {
		t.Error("fundingId empty")
	}
	return agreementID
}

func getTransferID(t *testing.T, response *httpexpect.Object) string {
	transferID := response.Value("data").Path("$.transfer_id").String().NotEmpty().Raw()
	if transferID == "" {
		t.Error("TransferId empty")
	}
	return transferID
}

func getTransactionID(t *testing.T, resp *httpexpect.Object) string {
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

func mintNFT(e *httpexpect.Expect, auth string, httpStatus int, payload map[string]interface{}) *httpexpect.Object {
	resp := addCommonHeaders(e.POST("/v1/nfts/registries/"+payload["registry_address"].(string)+"/mint"), auth).
		WithJSON(payload).
		Expect().Status(httpStatus)

	httpObj := resp.JSON().Object()
	return httpObj
}

func transferNFT(e *httpexpect.Expect, auth string, httpStatus int, payload map[string]interface{}) *httpexpect.Object {
	resp := addCommonHeaders(e.POST("/v1/nfts/registries/"+payload["registry_address"].(string)+"/tokens/"+payload["token_id"].(string)+"/transfer"), auth).
		WithJSON(payload).
		Expect().Status(httpStatus)

	httpObj := resp.JSON().Object()
	return httpObj
}

func ownerOfNFT(e *httpexpect.Expect, auth string, httpStatus int, payload map[string]interface{}) *httpexpect.Value {
	objGet := addCommonHeaders(e.GET("/v1/nfts/registries/"+payload["registry_address"].(string)+"/tokens/"+payload["token_id"].(string)+"/owner"), auth).
		Expect().Status(httpStatus).JSON().NotNull()
	return objGet
}

func getProof(e *httpexpect.Expect, auth string, httpStatus int, documentID string, payload map[string]interface{}) *httpexpect.Object {
	resp := addCommonHeaders(e.POST("/v1/documents/"+documentID+"/proofs"), auth).
		WithJSON(payload).
		Expect().Status(httpStatus)
	return resp.JSON().Object()
}

func getAccount(e *httpexpect.Expect, auth string, httpStatus int, identifier string) *httpexpect.Object {
	resp := addCommonHeaders(e.GET("/v1/accounts/"+identifier), auth).
		Expect().Status(httpStatus)
	return resp.JSON().Object()
}

func getAllAccounts(e *httpexpect.Expect, auth string, httpStatus int) *httpexpect.Object {
	resp := addCommonHeaders(e.GET("/v1/accounts"), auth).
		Expect().Status(httpStatus)
	return resp.JSON().Object()
}

func generateAccount(e *httpexpect.Expect, auth string, httpStatus int, payload map[string]map[string]string) *httpexpect.Object {
	resp := addCommonHeaders(e.POST("/v1/accounts/generate"), auth).
		WithJSON(payload).
		Expect().Status(httpStatus)
	return resp.JSON().Object()
}

// TODO add rest of the endpoints for config

func createInsecureClient() *http.Client {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	return &http.Client{Transport: tr}
}

func getTransactionStatusAndMessage(e *httpexpect.Expect, auth string, txID string) (string, string) {
	emptyResponseTolerance := 5
	emptyResponsesEncountered := 0
	for {
		resp := addCommonHeaders(e.GET("/v1/jobs/"+txID), auth).Expect().Status(200).JSON().Object().Raw()
		status, ok := resp["status"].(string)
		if !ok {
			emptyResponsesEncountered++
			if emptyResponsesEncountered > emptyResponseTolerance {
				panic("transaction api non-responsive")
			}
			time.Sleep(1 * time.Second)
			continue
		}

		if status == "pending" {
			time.Sleep(1 * time.Second)
			continue
		}

		message, ok := resp["message"].(string)

		if !ok {
			message = "Unknown error while processing transaction"
		}

		return status, message
	}
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

func getGenericDocumentAndCheck(t *testing.T, e *httpexpect.Expect, auth string, documentID string, params map[string]interface{}, attrs coreapi.AttributeMapRequest) *httpexpect.Value {
	objGet := addCommonHeaders(e.GET("/v1/documents/"+documentID), auth).
		Expect().Status(http.StatusOK).JSON().NotNull()
	objGet.Path("$.header.document_id").String().Equal(documentID)
	for k, v := range params {
		objGet.Path("$.data." + k).String().Equal(v.(string))
	}

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
	}
	return objGet
}

func getClonedDocumentAndCheck(t *testing.T, e *httpexpect.Expect, auth string, docID string, docID1 string, params map[string]interface{}, attrs coreapi.AttributeMapRequest) *httpexpect.Value {
	objGet := addCommonHeaders(e.GET("/v1/documents/"+docID), auth).
		Expect().Status(http.StatusOK).JSON().NotNull()
	objGet.Path("$.header.document_id").String().Equal(docID)
	for k, v := range params {
		objGet.Path("$.data." + k).String().Equal(v.(string))
	}

	objGet1 := addCommonHeaders(e.GET("/v1/documents/"+docID1), auth).
		Expect().Status(http.StatusOK).JSON().NotNull()
	objGet1.Path("$.header.document_id").String().Equal(docID1)
	for k, v := range params {
		objGet.Path("$.data." + k).String().Equal(v.(string))
	}

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
		fmt.Println("************************", objGet, objGet1)
	assert.Equal(t, gattrs, gattrs1)
	}
	return objGet
}

func nonExistingGenericDocumentCheck(e *httpexpect.Expect, auth string, documentID string) *httpexpect.Value {
	objGet := addCommonHeaders(e.GET("/v1/documents/"+documentID), auth).
		Expect().Status(404).JSON().NotNull()
	return objGet
}

func nonExistingGenericDocumentVersionCheck(e *httpexpect.Expect, auth string, documentID, versionID string) *httpexpect.Value {
	objGet := addCommonHeaders(e.GET("/v1/documents/"+documentID+"/versions/"+versionID), auth).
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

func addTransitionRules(e *httpexpect.Expect, auth, docID string, payload map[string][]map[string]string, status int) *httpexpect.Object {
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
