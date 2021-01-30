// +build integration unit testworld

package ideth

import (
	"context"
	"testing"
	"time"

	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/contextutil"
	"github.com/centrifuge/go-centrifuge/ethereum"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/jobs/jobsv2"
	testingconfig "github.com/centrifuge/go-centrifuge/testingutils/config"
	"github.com/centrifuge/gocelery/v2"
	"github.com/stretchr/testify/assert"
)

func (b *Bootstrapper) TestBootstrap(context map[string]interface{}) error {
	return b.Bootstrap(context)
}

func (*Bootstrapper) TestTearDown() error {
	return nil
}

func DeployIdentity(t *testing.T, ctx map[string]interface{}, cfg config.Configuration) identity.DID {
	factory := ctx[identity.BootstrappedDIDFactory].(identity.Factory)
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
	txn, err := factory.CreateIdentity(acc.GetEthereumDefaultAccountName(), ethKeys)
	assert.Nil(t, err, "send create identity should be successful")
	d := ctx[jobsv2.BootstrappedDispatcher].(jobsv2.Dispatcher)
	client := ctx[ethereum.BootstrappedEthereumClient].(ethereum.Client)
	ctxh := context.Background()
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
	return did
}
