//go:build testworld

package testworld

import (
	"net/http"
	"strings"
	"testing"

	testingcommons "github.com/centrifuge/go-centrifuge/testingutils/common"

	"github.com/centrifuge/go-centrifuge/testworld/park/behavior/client"
	"github.com/centrifuge/go-centrifuge/testworld/park/host"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/stretchr/testify/assert"
)

func TestDocumentsAPI_GetRoles(t *testing.T) {
	t.Parallel()

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

	aliceClient, err := controller.GetClientForHost(t, host.Alice)
	assert.NoError(t, err)

	bob, err := controller.GetHost(host.Bob)
	assert.NoError(t, err)

	// Alice prepares document to share with Bob
	payload := genericCoreAPICreate([]string{bob.GetMainAccount().GetAccountID().ToHexString()})

	res := aliceClient.CreateDocument("documents", http.StatusCreated, payload)

	status := client.GetDocumentStatus(res)
	assert.Equal(t, status, "pending")

	docID := client.GetDocumentIdentifier(res)

	for _, c := range tests {
		aliceClient.GetRole(docID, c.roleID, c.status)
	}
}

func TestDocumentsAPI_AddAndUpdate(t *testing.T) {
	t.Parallel()

	bob, err := controller.GetHost(host.Bob)
	assert.NoError(t, err)

	aliceClient, err := controller.GetClientForHost(t, host.Alice)
	assert.NoError(t, err)

	// Alice prepares document to share with Bob and charlie
	payload := genericCoreAPICreate([]string{bob.GetMainAccount().GetAccountID().ToHexString()})
	res := aliceClient.CreateDocument("documents", http.StatusCreated, payload)

	status := client.GetDocumentStatus(res)
	assert.Equal(t, status, "pending")

	docID := client.GetDocumentIdentifier(res)
	roleID := hexutil.Encode(utils.RandomSlice(32))

	// missing role
	aliceClient.GetRole(docID, roleID, http.StatusNotFound)

	// add role
	collab, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	obj := aliceClient.AddRole(docID, roleID, []string{collab.ToHexString()}, http.StatusOK)
	groleID, gcollabs := client.ParseRole(obj)

	assert.Equal(t, roleID, groleID)
	assert.Equal(t, []string{strings.ToLower(collab.ToHexString())}, gcollabs)

	aliceClient.GetRole(docID, roleID, http.StatusOK)

	// update role
	collab, err = testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	aliceClient.UpdateRole(docID, roleID, []string{collab.ToHexString()}, http.StatusOK)

	obj = aliceClient.GetRole(docID, roleID, http.StatusOK)

	groleID, gcollabs = client.ParseRole(obj)

	assert.Equal(t, roleID, groleID)
	assert.Equal(t, []string{strings.ToLower(collab.ToHexString())}, gcollabs)
}
