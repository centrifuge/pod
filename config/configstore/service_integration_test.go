// +build integration

package configstore_test

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/centrifuge/go-centrifuge/bootstrap/bootstrappers/testingbootstrap"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/identity"
)

var identityService identity.Service
var cfg config.Service

func TestMain(m *testing.M) {
	// Adding delay to startup (concurrency hack)
	time.Sleep(time.Second + 2)
	ctx := testingbootstrap.TestFunctionalEthereumBootstrap()
	cfg = ctx[config.BootstrappedConfigStorage].(config.Service)
	identityService = ctx[identity.BootstrappedIDService].(identity.Service)
	result := m.Run()
	testingbootstrap.TestFunctionalEthereumTearDown()
	os.Exit(result)
}

func TestService_GenerateTenantHappy(t *testing.T) {
	tct, err := cfg.GenerateTenant()
	assert.NoError(t, err)
	i, _ := tct.GetIdentityID()
	tc, err := cfg.GetTenant(i)
	assert.NoError(t, err)
	assert.NotNil(t, tc)
	i, _ = tc.GetIdentityID()
	cid, err := identity.ToCentID(i)
	assert.NoError(t, err)
	assert.True(t, tc.GetEthereumDefaultAccountName() != "")
	pb, pv := tc.GetSigningKeyPair()
	err = checkKeyPair(t, pb, pv)
	pb, pv = tc.GetEthAuthKeyPair()
	err = checkKeyPair(t, pb, pv)
	exists, err := identityService.CheckIdentityExists(cid)
	assert.NoError(t, err)
	assert.True(t, exists)
}

func checkKeyPair(t *testing.T, pb string, pv string) error {
	assert.True(t, pb != "")
	assert.True(t, pv != "")
	_, err := os.Stat(pb)
	assert.False(t, os.IsNotExist(err))
	_, err = os.Stat(pv)
	assert.False(t, os.IsNotExist(err))
	return err
}
