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

	cacc := map[string]string{
		"id":            "0xc81ebbec0559a6acf184535eb19da51ed3ed8c4ac65323999482aaf9b6696e27",
		"secret":        "0xc166b100911b1e9f780bb66d13badf2c1edbe94a1220f1a0584c09490158be31",
		"ss_58_address": "5Gb6Zfe8K8NSKrkFLCgqs8LUdk7wKweXM5pN296jVqDpdziR",
	}

	// generate a tenant within Charlie
	res = generateAccount(charlie.httpExpect, charlie.id.String(), http.StatusOK, cacc)
	tcID := res.Value("identity_id").String().NotEmpty()
	tcID.NotEqual(charlie.id.String())
}
