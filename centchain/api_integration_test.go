//go:build integration

package centchain_test

import (
	"context"
	"os"
	"testing"

	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/centrifuge/pod/bootstrap"
	"github.com/centrifuge/pod/bootstrap/bootstrappers/integration_test"
	"github.com/centrifuge/pod/bootstrap/bootstrappers/testlogging"
	"github.com/centrifuge/pod/centchain"
	"github.com/centrifuge/pod/config"
	"github.com/centrifuge/pod/config/configstore"
	"github.com/centrifuge/pod/contextutil"
	"github.com/centrifuge/pod/dispatcher"
	v2 "github.com/centrifuge/pod/identity/v2"
	"github.com/centrifuge/pod/jobs"
	"github.com/centrifuge/pod/pallets"
	"github.com/centrifuge/pod/storage/leveldb"
	genericUtils "github.com/centrifuge/pod/testingutils/generic"
	"github.com/centrifuge/pod/testingutils/keyrings"
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
	serviceCtx := bootstrap.RunTestBootstrappers(integrationTestBootstrappers, nil)
	testAPI = genericUtils.GetService[centchain.API](serviceCtx)
	cfgSrv = genericUtils.GetService[config.Service](serviceCtx)

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

func TestApi_SubmitAndWatch_ExtrinsicSuccess(t *testing.T) {
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

	_, err = testAPI.SubmitAndWatch(ctx, meta, call, podOperator.ToKeyringPair())
	assert.NoError(t, err)
}

func TestApi_GetPendingExtrinsics(t *testing.T) {
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

	_, _, _, err = testAPI.SubmitExtrinsic(ctx, meta, call, podOperator.ToKeyringPair())
	assert.NoError(t, err)

	pendingExtrinsics, err := testAPI.GetPendingExtrinsics()
	assert.NoError(t, err)

	assert.True(t, len(pendingExtrinsics) > 0)
}
