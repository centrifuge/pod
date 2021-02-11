// +build integration

package ethereum_test

import (
	"context"
	"fmt"
	"math/big"
	"os"
	"testing"

	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/bootstrap/bootstrappers/testlogging"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/config/configstore"
	"github.com/centrifuge/go-centrifuge/ethereum"
	"github.com/centrifuge/go-centrifuge/identity/ideth"
	"github.com/centrifuge/go-centrifuge/jobs"
	"github.com/centrifuge/go-centrifuge/storage/leveldb"
	"github.com/stretchr/testify/assert"
)

var cfg config.Configuration
var ctx = map[string]interface{}{}

func TestMain(m *testing.M) {
	var bootstappers = []bootstrap.TestBootstrapper{
		&testlogging.TestLoggingBootstrapper{},
		&config.Bootstrapper{},
		&leveldb.Bootstrapper{},
		jobs.Bootstrapper{},
		ethereum.Bootstrapper{},
		&ideth.Bootstrapper{},
		&configstore.Bootstrapper{},
	}

	bootstrap.RunTestBootstrappers(bootstappers, ctx)
	cfg = ctx[bootstrap.BootstrappedConfig].(config.Configuration)

	result := m.Run()
	bootstrap.RunTestTeardown(bootstappers)
	os.Exit(result)
}

func TestGetConnection_returnsSameConnection(t *testing.T) {
	t.Parallel()
	howMany := 5
	confChannel := make(chan ethereum.Client, howMany)
	for ix := 0; ix < howMany; ix++ {
		go func() {
			confChannel <- ethereum.GetClient()
		}()
	}
	for ix := 0; ix < howMany; ix++ {
		multiThreadCreatedCon := <-confChannel
		assert.Equal(t, multiThreadCreatedCon, ethereum.GetClient(), "Should only return a single ethereum client")
	}
}

func TestGethClient_NoEthKeyProvided(t *testing.T) {
	ethAcc, err := cfg.GetEthereumAccount("main")
	assert.NoError(t, err)
	cfg.Set("ethereum.accounts.main.key", "")
	_, err = ethereum.NewGethClient(cfg)
	assert.Error(t, err)
	assert.Equal(t, ethereum.ErrEthKeyNotProvided, err)
	cfg.Set("ethereum.accounts.main.key", ethAcc.Key)
}

func TestGethClient_GetTxOpts(t *testing.T) {
	// Environmental error in travis only, unskip when fixed
	t.SkipNow()
	cfg.Set("ethereum.maxGasPrice", 30000000000)
	gc, err := ethereum.NewGethClient(cfg)
	assert.NoError(t, err)
	assert.NotNil(t, gc)

	opts, err := gc.GetTxOpts(context.Background(), "main")
	assert.NoError(t, err)
	fmt.Println("Calculated GasPrice 1", opts.GasPrice.String())
	assert.True(t, opts.GasPrice.Cmp(big.NewInt(20000000000)) == 0)

	cfg.Set("ethereum.maxGasPrice", 10000000000)
	gc, err = ethereum.NewGethClient(cfg)
	assert.NoError(t, err)
	assert.NotNil(t, gc)
	opts, err = gc.GetTxOpts(context.Background(), "main")
	assert.NoError(t, err)
	fmt.Println("Calculated GasPrice 2", opts.GasPrice.String())
	assert.True(t, opts.GasPrice.Cmp(big.NewInt(10000000000)) == 0)
}

func TestGethClient_GetBlockByNumber_MaxRetries(t *testing.T) {
	gc, err := ethereum.NewGethClient(cfg)
	assert.NoError(t, err)
	assert.NotNil(t, gc)
	_, err = gc.GetBlockByNumber(context.Background(), big.NewInt(1000000000000000000))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Error retrying getting block number")
}

func BenchmarkGethClient_GetTxOpts(b *testing.B) {
	gc, err := ethereum.NewGethClient(cfg)
	assert.NoError(b, err)
	assert.NotNil(b, gc)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			opts, err := gc.GetTxOpts(context.Background(), "main")
			assert.NoError(b, err)
			assert.True(b, opts.GasPrice.Cmp(big.NewInt(20000000000)) == 0)
		}
	})
}
