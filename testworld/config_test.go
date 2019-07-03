// +build testworld

package testworld

import (
	"net/http"
	"strings"
	"testing"
)

func TestConfig_Happy(t *testing.T) {
	t.Parallel()
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

	// generate a tenant within Charlie
	res = generateAccount(charlie.httpExpect, charlie.id.String(), http.StatusOK)
	tcID := res.Value("identity_id").String().NotEmpty()
	tcID.NotEqual(charlie.id.String())
}
