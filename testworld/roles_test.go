//go:build testworld
// +build testworld

package testworld

import (
	"net/http"
	"strings"
	"testing"

	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/gavv/httpexpect"
	"github.com/stretchr/testify/assert"
)

func TestGetRoles_error(t *testing.T) {
	tests := []struct {
		roleID string
		status int
	}{
		{
			roleID: "some id",
			status: http.StatusBadRequest,
		},

		{
			roleID: hexutil.Encode(utils.RandomSlice(32)),
			status: http.StatusNotFound,
		},
	}

	alice := doctorFord.getHostTestSuite(t, "Alice")
	bob := doctorFord.getHostTestSuite(t, "Bob")

	// Alice prepares document to share with Bob and charlie
	payload := genericCoreAPICreate([]string{bob.id.String()})
	res := createDocument(alice.httpExpect, alice.id.String(), "documents", http.StatusCreated, payload)
	status := getDocumentStatus(t, res)
	assert.Equal(t, status, "pending")
	docID := getDocumentIdentifier(t, res)
	var obj *httpexpect.Object
	for _, c := range tests {
		obj = getRole(alice.httpExpect, alice.id.String(), docID, c.roleID, c.status)
	}

	// hack to ensure the tests aren't skipped
	_ = obj
}

func TestRoles_Add_Update(t *testing.T) {
	alice := doctorFord.getHostTestSuite(t, "Alice")
	bob := doctorFord.getHostTestSuite(t, "Bob")

	// Alice prepares document to share with Bob and charlie
	payload := genericCoreAPICreate([]string{bob.id.String()})
	res := createDocument(alice.httpExpect, alice.id.String(), "documents", http.StatusCreated, payload)
	status := getDocumentStatus(t, res)
	assert.Equal(t, status, "pending")
	docID := getDocumentIdentifier(t, res)
	roleID := hexutil.Encode(utils.RandomSlice(32))

	// missing role
	getRole(alice.httpExpect, alice.id.String(), docID, roleID, http.StatusNotFound)

	// add role
	collab := testingidentity.GenerateRandomDID()
	obj := addRole(alice.httpExpect, alice.id.String(), docID, roleID, []string{collab.String()}, http.StatusOK)
	groleID, gcollabs := parseRole(obj)
	assert.Equal(t, roleID, groleID)
	assert.Equal(t, []string{strings.ToLower(collab.String())}, gcollabs)
	getRole(alice.httpExpect, alice.id.String(), docID, roleID, http.StatusOK)

	// update role
	collab = testingidentity.GenerateRandomDID()
	updateRole(alice.httpExpect, alice.id.String(), docID, roleID, []string{collab.String()}, http.StatusOK)
	obj = getRole(alice.httpExpect, alice.id.String(), docID, roleID, http.StatusOK)
	groleID, gcollabs = parseRole(obj)
	assert.Equal(t, roleID, groleID)
	assert.Equal(t, []string{strings.ToLower(collab.String())}, gcollabs)
}
