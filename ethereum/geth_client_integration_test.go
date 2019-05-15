// +build integration

package ethereum_test

import (
	"context"
	"math/big"
	"os"
	"testing"
	"time"

	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/bootstrap/bootstrappers/testlogging"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/config/configstore"
	"github.com/centrifuge/go-centrifuge/ethereum"
	"github.com/centrifuge/go-centrifuge/identity/ideth"
	"github.com/centrifuge/go-centrifuge/jobs"
	"github.com/centrifuge/go-centrifuge/jobs/jobsv1"
	"github.com/centrifuge/go-centrifuge/queue"
	"github.com/centrifuge/go-centrifuge/storage/leveldb"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var cfg config.Configuration
var ctx = map[string]interface{}{}

func registerMockedTransactionTask() {
	queueSrv := ctx[bootstrap.BootstrappedQueueServer].(*queue.Server)
	jobManager := ctx[jobs.BootstrappedService].(jobs.Manager)

	mockClient := &ethereum.MockEthClient{}

	// txHash: 0x1 -> successful
	mockClient.On("TransactionByHash", mock.Anything, common.HexToHash("0x1")).Return(&types.Transaction{}, false, nil).Once()
	mockClient.On("TransactionReceipt", mock.Anything, common.HexToHash("0x1")).Return(&types.Receipt{Status: 1}, nil).Once()

	// txHash: 0x2 -> fail
	mockClient.On("TransactionByHash", mock.Anything, common.HexToHash("0x2")).Return(&types.Transaction{}, false, nil).Once()
	mockClient.On("TransactionReceipt", mock.Anything, common.HexToHash("0x2")).Return(&types.Receipt{Status: 0}, nil).Once()

	// txHash: 0x3 -> pending
	mockClient.On("TransactionByHash", mock.Anything, common.HexToHash("0x3")).Return(&types.Transaction{}, true, nil).Maybe()

	ethTransTask := ethereum.NewTransactionStatusTask(200*time.Millisecond, jobManager, mockClient.TransactionByHash, mockClient.TransactionReceipt, ethereum.DefaultWaitForTransactionMiningContext)
	queueSrv.RegisterTaskType(ethereum.EthTXStatusTaskName, ethTransTask)

}

func TestMain(m *testing.M) {
	var bootstappers = []bootstrap.TestBootstrapper{
		&testlogging.TestLoggingBootstrapper{},
		&config.Bootstrapper{},
		&leveldb.Bootstrapper{},
		jobsv1.Bootstrapper{},
		&queue.Bootstrapper{},
		ethereum.Bootstrapper{},
		&ideth.Bootstrapper{},
		&configstore.Bootstrapper{},
	}

	bootstrap.RunTestBootstrappers(bootstappers, ctx)
	cfg = ctx[bootstrap.BootstrappedConfig].(config.Configuration)

	queueStartBootstrap := &queue.Starter{}
	bootstappers = append(bootstappers, queueStartBootstrap)
	// register queue task
	registerMockedTransactionTask()
	//start queue
	queueStartBootstrap.TestBootstrap(ctx)

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
	cfg.Set("ethereum.maxGasPrice", 30000000000)
	gc, err := ethereum.NewGethClient(cfg)
	assert.NoError(t, err)
	assert.NotNil(t, gc)

	opts, err := gc.GetTxOpts(context.Background(), "main")
	assert.NoError(t, err)
	assert.True(t, opts.GasPrice.Cmp(big.NewInt(20000000000)) == 0)

	cfg.Set("ethereum.maxGasPrice", 10000000000)
	gc, err = ethereum.NewGethClient(cfg)
	assert.NoError(t, err)
	assert.NotNil(t, gc)
	opts, err = gc.GetTxOpts(context.Background(), "main")
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
