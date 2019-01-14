// +build testworld

package testworld

import (
	"net/http"
	"testing"

	"github.com/gavv/httpexpect"
)

func TestConfig_Happy(t *testing.T) {
	t.Parallel()
	charlie := doctorFord.getHostTestSuite(t, "Charlie")

	// check charlies node config
	res := getNodeConfig(charlie.httpExpect, charlie.id.String(), http.StatusOK)
	tenantID := res.Value("main_identity").Path("$.identity_id").String().NotEmpty()
	tenantID.Equal(charlie.id.String())

	// check charlies main tenant config
	res = getTenantConfig(charlie.httpExpect, charlie.id.String(), http.StatusOK, charlie.id.String())
	tenantID2 := res.Value("identity_id").String().NotEmpty()
	tenantID2.Equal(charlie.id.String())

	// check charlies all tenant configs
	res = getAllTenantConfigs(charlie.httpExpect, charlie.id.String(), http.StatusOK)
	tenants := res.Value("data").Array()
	tids := getAccounts(tenants)
	if _, ok := tids[charlie.id.String()]; !ok {
		t.Error("Charlies id needs to exist in the accounts list")
	}

	// generate a tenant within Charlie
	res = generateTenant(charlie.httpExpect, charlie.id.String(), http.StatusOK)
	tcID := res.Value("identity_id").String().NotEmpty()
	tcID.NotEqual(charlie.id.String())
}

func getAccounts(accounts *httpexpect.Array) map[string]string {
	tids := make(map[string]string)
	for i := 0; i < int(accounts.Length().Raw()); i++ {
		val := accounts.Element(i).Path("$.identity_id").String().NotEmpty().Raw()
		tids[val] = val
	}
	return tids
}
