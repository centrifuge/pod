// +build testworld

package testworld

import (
	"net/http"
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
	res := createDocumentV2(alice.httpExpect, alice.id.String(), "documents", http.StatusCreated, payload)
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
