// +build testworld

package testworld

import (
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfig_Happy(t *testing.T) {
	charlie := doctorFord.getHostTestSuite(t, "Charlie")

	// check charlies main account
	res := getAccount(charlie.httpExpect, charlie.id.String(), http.StatusOK, charlie.id.String())
	accountID2 := res.Value("identity_id").String().NotEmpty()
	accountID2.Equal(strings.ToLower(charlie.id.String()))

	// check charlies all accounts
	res = getAllAccounts(charlie.httpExpect, charlie.id.String(), http.StatusOK)
	tenants := res.Value("data").Array()
	accIDs := getAccounts(tenants)
	if _, ok := accIDs[strings.ToLower(charlie.id.String())]; !ok {
		t.Error("Charlies id needs to exist in the accounts list")
	}

	cacc := map[string]map[string]string{
		"centrifuge_chain_account": {
			"id":            "0xd43593c715fdd31c61141abd04a99fd6822c8558854ccde39a5684e7a56da27d",
			"secret":        "//Alice",
			"ss_58_address": "5GrwvaEF5zXb26Fz9rcQpDWS57CtERHpNehXCPcNoHGKutQY",
		},
	}

	// generate a tenant within Charlie
	did, err := generateAccount(doctorFord.maeve, charlie.httpExpect, charlie.id.String(), http.StatusCreated, cacc)
	assert.NoError(t, err)
	assert.False(t, did.Equal(charlie.id))
}
