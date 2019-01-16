// +build testworld

package testworld

import (
	"net/http"
	"testing"
)

func TestConfig_Happy(t *testing.T) {
	t.Parallel()
	charlie := doctorFord.getHostTestSuite(t, "Charlie")

	// check charlies node config
	res := getNodeConfig(charlie.httpExpect, charlie.id.String(), http.StatusOK)
	tenantID := res.Value("main_identity").Path("$.identity_id").String().NotEmpty()
	tenantID.Equal(charlie.id.String())

	// check charlies main tenant config
	res = getAccount(charlie.httpExpect, charlie.id.String(), http.StatusOK, charlie.id.String())
	tenantID2 := res.Value("identity_id").String().NotEmpty()
	tenantID2.Equal(charlie.id.String())

	// check charlies all tenant configs
	res = getAllAccounts(charlie.httpExpect, charlie.id.String(), http.StatusOK)
	tenants := res.Value("data").Array()
	tids := getAccounts(tenants)
	if _, ok := tids[charlie.id.String()]; !ok {
		t.Error("Charlies id needs to exist in the accounts list")
	}

	// generate a tenant within Charlie
	res = generateAccount(charlie.httpExpect, charlie.id.String(), http.StatusOK)
	tcID := res.Value("identity_id").String().NotEmpty()
	tcID.NotEqual(charlie.id.String())
}
