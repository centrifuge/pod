// +build integration

package ideth

import (
	"context"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/bootstrap/bootstrappers/testlogging"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/config/configstore"
	"github.com/centrifuge/go-centrifuge/contextutil"
	"github.com/centrifuge/go-centrifuge/ethereum"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/jobs/jobsv1"
	"github.com/centrifuge/go-centrifuge/jobs/jobsv2"
	"github.com/centrifuge/go-centrifuge/queue"
	"github.com/centrifuge/go-centrifuge/storage/leveldb"
	"github.com/centrifuge/go-centrifuge/testingutils"
	testingconfig "github.com/centrifuge/go-centrifuge/testingutils/config"
	"github.com/centrifuge/gocelery/v2"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
)

var cfg config.Configuration
var ctx = map[string]interface{}{}

func TestMain(m *testing.M) {
	ctx = testingutils.BuildIntegrationTestingContext()
	var bootstappers = []bootstrap.TestBootstrapper{
		&testlogging.TestLoggingBootstrapper{},
		&config.Bootstrapper{},
		&leveldb.Bootstrapper{},
		jobsv1.Bootstrapper{},
		jobsv2.Bootstrapper{},
		&queue.Bootstrapper{},
		ethereum.Bootstrapper{},
		&Bootstrapper{},
		&configstore.Bootstrapper{},
		&Bootstrapper{},
		&queue.Starter{},
	}

	bootstrap.RunTestBootstrappers(bootstappers, ctx)
	cfg = ctx[bootstrap.BootstrappedConfig].(config.Configuration)
	result := m.Run()
	bootstrap.RunTestTeardown(bootstappers)
	os.Exit(result)
}

func TestCreateIdentity_successful(t *testing.T) {
	factory := ctx[identity.BootstrappedDIDFactoryV2].(identity.FactoryInterface)
	accountCtx := testingconfig.CreateAccountContext(t, cfg)
	acc, err := contextutil.Account(accountCtx)
	assert.NoError(t, err)
	did, err := factory.NextIdentityAddress()
	assert.NoError(t, err)
	exists, err := factory.IdentityExists(did)
	assert.NoError(t, err)
	assert.False(t, exists)
	keys, err := acc.GetKeys()
	assert.NoError(t, err)
	ethKeys, err := identity.ConvertAccountKeysToKeyDID(keys)
	assert.NoError(t, err)
	txn, err := factory.CreateIdentity(
		acc.GetEthereumDefaultAccountName(),
		common.HexToAddress(acc.GetEthereumAccount().Address), ethKeys)
	assert.Nil(t, err, "send create identity should be successful")
	d := ctx[jobsv2.BootstrappedDispatcher].(jobsv2.Dispatcher)
	client := ctx[ethereum.BootstrappedEthereumClient].(ethereum.Client)
	ctxh, canc := context.WithCancel(context.Background())
	defer canc()
	wg := sync.WaitGroup{}
	wg.Add(1)
	errOut := make(chan error)
	go d.Start(ctxh, &wg, errOut)
	ok := d.RegisterRunnerFunc("ethWait", func([]interface{}, map[string]interface{}) (result interface{}, err error) {
		return ethereum.IsTxnSuccessful(ctxh, client, txn.Hash())
	})
	assert.True(t, ok)
	job := gocelery.NewRunnerFuncJob("Eth wait", "ethWait", nil, nil, time.Time{})
	res, err := d.Dispatch(did, job)
	assert.NoError(t, err)
	_, err = res.Await(ctxh)
	assert.NoError(t, err)
	exists, err = factory.IdentityExists(did)
	assert.NoError(t, err)
	assert.True(t, exists)
	contractCode, err := client.GetEthClient().CodeAt(context.Background(), did.ToAddress(), nil)
	assert.Nil(t, err, "should be successful to get the contract code")
	assert.Equal(t, true, len(contractCode) > 3000, "current contract code should be around 3378 bytes")
}
