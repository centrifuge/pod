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
	res := getNodeConfig(charlie.httpExpect, http.StatusOK)
	tenantID := res.Value("main_identity").Path("$.identity_id").String().NotEmpty()
	tenantID.Equal(charlie.id.String())

	// check charlies main tenant config
	res = getTenantConfig(charlie.httpExpect, http.StatusOK, charlie.id.String())
	tenantID2 := res.Value("identity_id").String().NotEmpty()
	tenantID2.Equal(charlie.id.String())

	// check charlies all tenant configs
	res = getAllTenantConfigs(charlie.httpExpect, http.StatusOK)
	tenants := res.Value("data").Array()
	tenants.Length().Equal(1)
	tenants.Element(0).Path("$.identity_id").String().NotEmpty().Equal(charlie.id.String())
}
