//go:build testworld

package client

import (
	"encoding/json"
	"testing"

	v2 "github.com/centrifuge/pod/http/v2"
	"github.com/gavv/httpexpect"
	"github.com/stretchr/testify/assert"
)

func (c *Client) GetRole(docID, roleID string, status int) *httpexpect.Object {
	objPost := addCommonHeaders(c.expect.GET("/v2/documents/"+docID+"/roles/"+roleID), c.jwtToken).
		Expect().Status(status).JSON().Object()
	return objPost
}

func (c *Client) AddRole(docID, roleID string, collaborators []string, status int) *httpexpect.Object {
	objPost := addCommonHeaders(c.expect.POST("/v2/documents/"+docID+"/roles"), c.jwtToken).WithJSON(map[string]interface{}{
		"key":           roleID,
		"collaborators": collaborators,
	}).Expect().Status(status).JSON().Object()
	return objPost
}

func (c *Client) UpdateRole(docID, roleID string, collaborators []string, status int) *httpexpect.Object {
	objPost := addCommonHeaders(c.expect.PATCH("/v2/documents/"+docID+"/roles/"+roleID), c.jwtToken).WithJSON(map[string]interface{}{
		"collaborators": collaborators,
	}).Expect().Status(status).JSON().Object()
	return objPost
}

func (c *Client) AddTransitionRules(docID string, payload map[string][]map[string]interface{}, status int) *httpexpect.Object {
	objPost := addCommonHeaders(c.expect.POST("/v2/documents/"+docID+"/transition_rules"), c.jwtToken).WithJSON(
		payload).Expect().Status(status).JSON().Object()
	return objPost
}

func (c *Client) GetTransitionRule(docID, ruleID string, status int) *httpexpect.Object {
	objPost := addCommonHeaders(c.expect.GET("/v2/documents/"+docID+"/transition_rules/"+ruleID), c.jwtToken).
		Expect().Status(status).JSON().Object()
	return objPost
}

func (c *Client) DeleteTransitionRule(docID, ruleID string, status int) *httpexpect.Response {
	objPost := addCommonHeaders(c.expect.DELETE("/v2/documents/"+docID+"/transition_rules/"+ruleID), c.jwtToken).
		Expect().Status(status)
	return objPost
}

func ParseRules(t *testing.T, obj *httpexpect.Object) v2.TransitionRules {
	d, err := json.Marshal(obj.Raw())
	assert.NoError(t, err)

	var tr v2.TransitionRules
	assert.NoError(t, json.Unmarshal(d, &tr))

	return tr
}

func ParseRule(t *testing.T, obj *httpexpect.Object) v2.TransitionRule {
	d, err := json.Marshal(obj.Raw())
	assert.NoError(t, err)

	var tr v2.TransitionRule
	assert.NoError(t, json.Unmarshal(d, &tr))

	return tr
}

func ParseRole(obj *httpexpect.Object) (roleID string, collaborators []string) {
	roleID = obj.Path("$.id").String().Raw()
	for _, c := range obj.Path("$.collaborators").Array().Iter() {
		collaborators = append(collaborators, c.String().Raw())
	}
	return roleID, collaborators
}
