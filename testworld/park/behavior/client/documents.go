//go:build testworld

package client

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/centrifuge/pod/http/coreapi"
	"github.com/stretchr/testify/assert"

	"github.com/gavv/httpexpect"
)

func (c *Client) CreateAndCommitDocument(payload map[string]interface{}) (string, error) {
	res := c.CreateDocument("documents", http.StatusCreated, payload)
	docID := GetDocumentIdentifier(res)
	res = c.CommitDocument("documents", http.StatusAccepted, docID)
	jobID, err := GetJobID(res)

	if err != nil {
		return "", err
	}

	if err = c.WaitForJobCompletion(jobID); err != nil {
		return "", err
	}

	return docID, err
}

func (c *Client) GetEntityRelationships(docIdentifier string) *httpexpect.Value {
	objGet := addCommonHeaders(c.expect.GET("/v2/entities/"+docIdentifier+"/relationships"), c.authToken).
		Expect().Status(http.StatusOK).JSON().NotNull()

	return objGet
}

func (c *Client) GetEntityWithRelation(relationshipID string) *httpexpect.Value {
	objGet := addCommonHeaders(c.expect.GET("/v2/relationships/"+relationshipID+"/entity"), c.authToken).
		Expect().Status(http.StatusOK).JSON().NotNull()

	return objGet
}

func (c *Client) NonexistentEntityWithRelation(relationshipID string) *httpexpect.Value {
	objGet := addCommonHeaders(c.expect.GET("/v2/relationships/"+relationshipID+"/entity"), c.authToken).
		Expect().Status(http.StatusNotFound).JSON().NotNull()

	return objGet
}

func (c *Client) NonExistingDocumentCheck(docIdentifier string) *httpexpect.Value {
	objGet := addCommonHeaders(c.expect.GET("/v2/documents/"+docIdentifier+"/committed"), c.authToken).
		Expect().Status(http.StatusNotFound).JSON().NotNull()
	return objGet
}

func (c *Client) NonExistingDocumentVersionCheck(docID, versionID string) *httpexpect.Value {
	objGet := addCommonHeaders(c.expect.GET("/v2/documents/"+docID+"/versions/"+versionID), c.authToken).
		Expect().Status(http.StatusNotFound).JSON().NotNull()
	return objGet
}

func (c *Client) CreateDocument(documentType string, status int, payload map[string]interface{}) *httpexpect.Object {
	obj := addCommonHeaders(c.expect.POST("/v2/"+documentType), c.authToken).
		WithJSON(payload).
		Expect().Status(status).JSON().Object()
	return obj
}

func (c *Client) CloneDocument(documentType string, status int, payload map[string]interface{}) *httpexpect.Object {
	obj := addCommonHeaders(c.expect.POST("/v2/"+documentType+"/"+payload["document_id"].(string)+"/clone"), c.authToken).
		WithJSON(payload).
		Expect().Status(status).JSON().Object()
	return obj
}

func (c *Client) UpdateDocument(documentType string, status int, payload map[string]interface{}) *httpexpect.Object {
	obj := addCommonHeaders(c.expect.PATCH("/v2/"+documentType+"/"+payload["document_id"].(string)), c.authToken).
		WithJSON(payload).
		Expect().Status(status).JSON().Object()
	return obj
}

func (c *Client) RemoveCollaborators(docType string, status int, docID string, collabs ...string) *httpexpect.Object {
	obj := addCommonHeaders(c.expect.DELETE("/v2/"+docType+"/"+docID+"/collaborators"), c.authToken).
		WithJSON(map[string][]string{
			"collaborators": collabs,
		}).
		Expect().Status(status).JSON().Object()
	return obj
}

func (c *Client) CommitDocument(documentType string, status int, docIdentifier string) *httpexpect.Object {
	obj := addCommonHeaders(c.expect.POST("/v2/"+documentType+"/"+docIdentifier+"/commit"), c.authToken).
		Expect().Status(status).JSON().Object()
	return obj
}

func (c *Client) GetProof(httpStatus int, documentID string, payload map[string]interface{}) *httpexpect.Object {
	resp := addCommonHeaders(c.expect.POST("/v2/documents/"+documentID+"/proofs"), c.authToken).
		WithJSON(payload).
		Expect().Status(httpStatus)
	return resp.JSON().Object()
}

func (c *Client) GetFingerprint(documentID string) string {
	obj := c.GetDocumentAndVerify(documentID, nil, nil)
	return obj.Path("$.header.fingerprint").String().Raw()
}

func (c *Client) GetDocumentAndVerify(documentID string, params map[string]interface{}, attrs coreapi.AttributeMapRequest) *httpexpect.Value {
	objGet := addCommonHeaders(c.expect.GET("/v2/documents/"+documentID+"/committed"), c.authToken).
		Expect().Status(http.StatusOK).JSON().NotNull()

	objGet.Path("$.header.document_id").String().Equal(documentID)
	for k, v := range params {
		objGet.Path("$.data." + k).String().Equal(v.(string))
	}

	if len(attrs) > 0 {
		reqJSON, err := json.Marshal(attrs)
		if err != nil {
			assert.Fail(c.t, err.Error())
		}

		gattrs := objGet.Path("$.attributes").Object().Raw()
		// Since we want to perform an equals check on the request attributes and response attributes we need to marshal and
		// unmarshal twice over the object
		respJSON, err := json.Marshal(gattrs)
		if err != nil {
			assert.Fail(c.t, err.Error())
		}
		var cattrs coreapi.AttributeMapRequest
		err = json.Unmarshal(respJSON, &cattrs)
		if err != nil {
			assert.Fail(c.t, err.Error())
		}
		respJSON, err = json.Marshal(cattrs)
		if err != nil {
			assert.Fail(c.t, err.Error())
		}

		assert.Equal(c.t, reqJSON, respJSON)
	}
	return objGet
}

func (c *Client) GetClonedDocumentAndCheck(docID string, docID1 string, params map[string]interface{}, attrs coreapi.AttributeMapRequest) *httpexpect.Value {
	objGet := addCommonHeaders(c.expect.GET("/v2/documents/"+docID+"/committed"), c.authToken).
		Expect().Status(http.StatusOK).JSON().NotNull()

	objGet.Path("$.header.document_id").String().Equal(docID)
	for k, v := range params {
		objGet.Path("$.data." + k).String().Equal(v.(string))
	}

	objGet1 := addCommonHeaders(c.expect.GET("/v2/documents/"+docID1+"/committed"), c.authToken).
		Expect().Status(http.StatusOK).JSON().NotNull()

	objGet1.Path("$.header.document_id").String().Equal(docID1)
	for k, v := range params {
		objGet.Path("$.data." + k).String().Equal(v.(string))
	}

	// make sure the fingerprints of the two documents are the same
	assert.Equal(c.t, objGet.Path("$.header.fingerprint"), objGet1.Path("$.header.fingerprint"))

	if len(attrs) > 0 {
		reqJson, err := json.Marshal(attrs)
		if err != nil {
			assert.Fail(c.t, err.Error())
		}

		gattrs := objGet.Path("$.attributes").Object().Raw()
		// Since we want to perform an equals check on the request attributes and response attributes we need to marshal and
		// unmarshal twice over the object
		respJson, err := json.Marshal(gattrs)
		if err != nil {
			assert.Fail(c.t, err.Error())
		}
		var cattrs coreapi.AttributeMapRequest
		err = json.Unmarshal(respJson, &cattrs)
		if err != nil {
			assert.Fail(c.t, err.Error())
		}
		respJson, err = json.Marshal(cattrs)
		if err != nil {
			assert.Fail(c.t, err.Error())
		}

		assert.Equal(c.t, reqJson, respJson)

		gattrs1 := objGet1.Path("$.attributes").Object().Raw()
		// Since we want to perform an equals check on the request attributes and response attributes we need to marshal and
		// unmarshal twice over the object
		respJson1, err := json.Marshal(gattrs1)
		if err != nil {
			assert.Fail(c.t, err.Error())
		}
		var cattrs1 coreapi.AttributeMapRequest
		err = json.Unmarshal(respJson1, &cattrs1)
		if err != nil {
			assert.Fail(c.t, err.Error())
		}
		respJson1, err = json.Marshal(cattrs1)
		if err != nil {
			assert.Fail(c.t, err.Error())
		}
		assert.Equal(c.t, gattrs, gattrs1)
	}
	return objGet
}

func (c *Client) NonExistingGenericDocumentVersionCheck(documentID, versionID string) *httpexpect.Value {
	objGet := addCommonHeaders(c.expect.GET("/v2/documents/"+documentID+"/versions/"+versionID), c.authToken).
		Expect().Status(404).JSON().NotNull()
	return objGet
}

func (c *Client) GetDocumentWithStatus(docID, status string, code int) *httpexpect.Value {
	objGet := addCommonHeaders(c.expect.GET("/v2/documents/"+docID+"/"+status), c.authToken).
		Expect().Status(code).JSON().NotNull()
	return objGet
}

func (c *Client) AddSignedAttribute(docID, label, payload, valType string) *httpexpect.Object {
	objPost := addCommonHeaders(c.expect.POST("/v2/documents/"+docID+"/signed_attribute"), c.authToken).WithJSON(map[string]string{
		"label":   label,
		"type":    valType,
		"payload": payload,
	}).Expect().Status(http.StatusOK).JSON().Object()
	return objPost
}

func SignedAttributeExists(t *testing.T, res *httpexpect.Object, label string) {
	attrs := res.Path("$.attributes").Object().Raw()
	_, ok := attrs[label]
	assert.True(t, ok)
}

func SignedAttributeMissing(t *testing.T, res *httpexpect.Object, label string) {
	attrs := res.Path("$.attributes").Object().Raw()
	_, ok := attrs[label]
	assert.False(t, ok)
}

func GetDocumentIdentifier(response *httpexpect.Object) string {
	docIdentifier := response.Value("header").Path("$.document_id").String().NotEmpty().Raw()
	return docIdentifier
}

func GetDocumentStatus(response *httpexpect.Object) string {
	status := response.Value("header").Path("$.status").String().NotEmpty().Raw()
	return status
}

func GetDocumentCurrentVersion(resp *httpexpect.Object) string {
	versionID := resp.Value("header").Path("$.version_id").String().Raw()

	return versionID
}

func CheckDocumentParams(obj *httpexpect.Object, params map[string]string) {
	for k, v := range params {
		obj.Path("$.data." + k).String().Equal(v)
	}
}
