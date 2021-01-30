// +build integration

package configstore_test

import (
	"context"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/bootstrap/bootstrappers/testingbootstrap"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/jobs/jobsv2"
	testingconfig "github.com/centrifuge/go-centrifuge/testingutils/config"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
)

var identityService identity.Service
var cfgSvc config.Service
var cfg config.Configuration
var dispatcher jobsv2.Dispatcher

type MockProtocolSetter struct{}

func (MockProtocolSetter) InitProtocolForDID(identity.DID) {
	// do nothing
}

func TestMain(m *testing.M) {
	// Adding delay to startup (concurrency hack)
	time.Sleep(time.Second + 2)
	ctx := testingbootstrap.TestFunctionalEthereumBootstrap()
	cfgSvc = ctx[config.BootstrappedConfigStorage].(config.Service)
	cfg = ctx[bootstrap.BootstrappedConfig].(config.Configuration)
	dispatcher = ctx[jobsv2.BootstrappedDispatcher].(jobsv2.Dispatcher)
	wg := new(sync.WaitGroup)
	wg.Add(1)
	ctxh, canc := context.WithCancel(context.Background())
	go dispatcher.Start(ctxh, wg, nil)
	ctx[bootstrap.BootstrappedPeer] = &MockProtocolSetter{}
	identityService = ctx[identity.BootstrappedDIDService].(identity.Service)
	result := m.Run()
	testingbootstrap.TestFunctionalEthereumTearDown()
	canc()
	wg.Wait()
	os.Exit(result)
}

func TestService_GenerateAccountHappy(t *testing.T) {
	// missing cent chain account
	didb, jobID, err := cfgSvc.GenerateAccountAsync(config.CentChainAccount{})
	assert.Error(t, err)

	// success
	didb, jobID, err = cfgSvc.GenerateAccountAsync(config.CentChainAccount{
		ID:       "0xc81ebbec0559a6acf184535eb19da51ed3ed8c4ac65323999482aaf9b6696e27",
		Secret:   "0xc166b100911b1e9f780bb66d13badf2c1edbe94a1220f1a0584c09490158be31",
		SS58Addr: "5Gb6Zfe8K8NSKrkFLCgqs8LUdk7wKweXM5pN296jVqDpdziR",
	})
	did := identity.NewDID(common.BytesToAddress(didb))
	assert.NoError(t, err)
	res, err := dispatcher.Result(did, jobID)
	assert.NoError(t, err)
	_, err = res.Await(context.Background())
	assert.NoError(t, err)
	tc, err := cfgSvc.GetAccount(did[:])
	assert.NoError(t, err)
	assert.NotNil(t, tc)
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
