// +build integration

package configstore_test

import (
	"github.com/centrifuge/go-centrifuge/testingutils/config"
	"os"
	"testing"
	"time"

	"github.com/centrifuge/go-centrifuge/bootstrap"

	"github.com/stretchr/testify/assert"

	"github.com/centrifuge/go-centrifuge/bootstrap/bootstrappers/testingbootstrap"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/identity"
)

var identityService identity.ServiceDID
var cfg config.Service

type MockProtocolSetter struct{}

func (MockProtocolSetter) InitProtocolForDID(DID *identity.DID) {
	// do nothing
}

func TestMain(m *testing.M) {
	// Adding delay to startup (concurrency hack)
	time.Sleep(time.Second + 2)
	ctx := testingbootstrap.TestFunctionalEthereumBootstrap()
	cfg = ctx[config.BootstrappedConfigStorage].(config.Service)
	ctx[bootstrap.BootstrappedPeer] = &MockProtocolSetter{}
	identityService = ctx[identity.BootstrappedDIDService].(identity.ServiceDID)
	result := m.Run()
	testingbootstrap.TestFunctionalEthereumTearDown()
	os.Exit(result)
}

func TestService_GenerateAccountHappy(t *testing.T) {
	tct, err := cfg.GenerateAccount()
	assert.NoError(t, err)
	i, _ := tct.GetIdentityID()
	tc, err := cfg.GetAccount(i)
	assert.NoError(t, err)
	assert.NotNil(t, tc)
	i, _ = tc.GetIdentityID()
	did := identity.NewDIDFromBytes(i)
	assert.True(t, tc.GetEthereumDefaultAccountName() != "")
	pb, pv := tc.GetSigningKeyPair()
	err = checkKeyPair(t, pb, pv)
	pb, pv = tc.GetEthAuthKeyPair()
	err = checkKeyPair(t, pb, pv)
	ctxh := testingconfig.CreateAccountContext(t, cfg.(config.Configuration))
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
