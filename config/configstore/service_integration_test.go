// +build integration

package configstore_test

import (
	"os"
	"testing"
	"time"

	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/bootstrap/bootstrappers/testingbootstrap"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/testingutils/config"
	"github.com/stretchr/testify/assert"
)

var identityService identity.Service
var cfgSvc config.Service
var cfg config.Configuration

type MockProtocolSetter struct{}

func (MockProtocolSetter) InitProtocolForDID(DID *identity.DID) {
	// do nothing
}

func TestMain(m *testing.M) {
	// Adding delay to startup (concurrency hack)
	time.Sleep(time.Second + 2)
	ctx := testingbootstrap.TestFunctionalEthereumBootstrap()
	cfgSvc = ctx[config.BootstrappedConfigStorage].(config.Service)
	cfg = ctx[bootstrap.BootstrappedConfig].(config.Configuration)
	ctx[bootstrap.BootstrappedPeer] = &MockProtocolSetter{}
	identityService = ctx[identity.BootstrappedDIDService].(identity.Service)
	result := m.Run()
	testingbootstrap.TestFunctionalEthereumTearDown()
	os.Exit(result)
}

func TestService_GenerateAccountHappy(t *testing.T) {
	tct, err := cfgSvc.GenerateAccount()
	assert.NoError(t, err)
	i := tct.GetIdentityID()
	tc, err := cfgSvc.GetAccount(i)
	assert.NoError(t, err)
	assert.NotNil(t, tc)
	i = tc.GetIdentityID()
	did, err := identity.NewDIDFromBytes(i)
	assert.NoError(t, err)
	assert.True(t, tc.GetEthereumDefaultAccountName() != "")
	pb, pv := tc.GetSigningKeyPair()
	err = checkKeyPair(t, pb, pv)
	ctxh := testingconfig.CreateAccountContext(t, cfg)
	err = identityService.Exists(ctxh, did)
	assert.NoError(t, err)
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
