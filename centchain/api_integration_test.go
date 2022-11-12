//go:build integration

package centchain_test

import (
	"context"
	"os"
	"testing"

	"github.com/centrifuge/go-centrifuge/pallets"

	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/bootstrap/bootstrappers/integration_test"
	"github.com/centrifuge/go-centrifuge/bootstrap/bootstrappers/testlogging"
	"github.com/centrifuge/go-centrifuge/centchain"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/config/configstore"
	"github.com/centrifuge/go-centrifuge/contextutil"
	"github.com/centrifuge/go-centrifuge/dispatcher"
	v2 "github.com/centrifuge/go-centrifuge/identity/v2"
	"github.com/centrifuge/go-centrifuge/jobs"
	"github.com/centrifuge/go-centrifuge/storage/leveldb"
	"github.com/centrifuge/go-centrifuge/testingutils/keyrings"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/stretchr/testify/assert"
)

var integrationTestBootstrappers = []bootstrap.TestBootstrapper{
	&integration_test.Bootstrapper{},
	&testlogging.TestLoggingBootstrapper{},
	&config.Bootstrapper{},
	&leveldb.Bootstrapper{},
	&configstore.Bootstrapper{},
	&jobs.Bootstrapper{},
	centchain.Bootstrapper{},
	&pallets.Bootstrapper{},
	&dispatcher.Bootstrapper{},
	&v2.AccountTestBootstrapper{},
}

var testAPI centchain.API
var cfgSrv config.Service

func TestMain(m *testing.M) {
	ctx := bootstrap.RunTestBootstrappers(integrationTestBootstrappers, nil)
	testAPI = ctx[centchain.BootstrappedCentChainClient].(centchain.API)
	cfgSrv = ctx[config.BootstrappedConfigStorage].(config.Service)

	result := m.Run()

	bootstrap.RunTestTeardown(integrationTestBootstrappers)

	os.Exit(result)
}

func TestApi_Call(t *testing.T) {
	var hash types.Hash
	err := testAPI.Call(&hash, "chain_getFinalizedHead")
	assert.NoError(t, err)
	assert.NotEmpty(t, hash.Hex())
}

func TestApi_GetMetadataLatest(t *testing.T) {
	meta, err := testAPI.GetMetadataLatest()
	assert.NoError(t, err)
	assert.NotNil(t, meta)
}

func TestApi_SubmitExtrinsic(t *testing.T) {
	meta, err := testAPI.GetMetadataLatest()
	assert.NoError(t, err)

	call, err := types.NewCall(meta, "System.remark", []byte{})
	assert.NoError(t, err)

	accounts, err := cfgSrv.GetAccounts()
	assert.NoError(t, err)

	acc := accounts[0]

	ctx := contextutil.WithAccount(context.Background(), acc)

	txHash, bn, sig, err := testAPI.SubmitExtrinsic(ctx, meta, call, keyrings.BobKeyRingPair)
	assert.NoError(t, err)
	assert.NotEmpty(t, txHash.Hex())
	assert.NotZero(t, bn)
	assert.True(t, sig.IsSr25519)
}

func TestApi_GetStorageLatest(t *testing.T) {
	meta, err := testAPI.GetMetadataLatest()
	assert.NoError(t, err)

	accounts, err := cfgSrv.GetAccounts()
	assert.NoError(t, err)

	acc := accounts[0]

	accountStorageKey, err := types.CreateStorageKey(meta, "System", "Account", acc.GetIdentity().ToBytes())
	assert.NoError(t, err)

	var accountInfo types.AccountInfo

	ok, err := testAPI.GetStorageLatest(accountStorageKey, &accountInfo)
	assert.NoError(t, err)
	assert.True(t, ok)
}

func TestApi_GetBlockLatest(t *testing.T) {
	block, err := testAPI.GetBlockLatest()
	assert.NoError(t, err)
	assert.NotNil(t, block)
}

func TestApi_SubmitAndWatch(t *testing.T) {
	meta, err := testAPI.GetMetadataLatest()
	assert.NoError(t, err)

	call, err := types.NewCall(meta, "System.remark", []byte{})
	assert.NoError(t, err)

	accounts, err := cfgSrv.GetAccounts()
	assert.NoError(t, err)

	acc := accounts[0]

	ctx := contextutil.WithAccount(context.Background(), acc)

	podOperator, err := cfgSrv.GetPodOperator()
	assert.NoError(t, err)

	info, err := testAPI.SubmitAndWatch(ctx, meta, call, podOperator.ToKeyringPair())
	assert.NoError(t, err)

	events, err := info.Events(meta)
	assert.NoError(t, err)
	assert.True(t, len(events.System_ExtrinsicSuccess) > 1)
}
