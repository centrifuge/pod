//go:build integration

package proxy_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/centrifuge/go-centrifuge/contextutil"

	v2 "github.com/centrifuge/go-centrifuge/identity/v2"

	keystoreType "github.com/centrifuge/chain-custom-types/pkg/keystore"
	proxyTypes "github.com/centrifuge/chain-custom-types/pkg/proxy"
	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/bootstrap/bootstrappers/integration_test"
	"github.com/centrifuge/go-centrifuge/bootstrap/bootstrappers/testlogging"
	"github.com/centrifuge/go-centrifuge/centchain"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/config/configstore"
	"github.com/centrifuge/go-centrifuge/dispatcher"
	"github.com/centrifuge/go-centrifuge/jobs"
	"github.com/centrifuge/go-centrifuge/pallets"
	"github.com/centrifuge/go-centrifuge/pallets/keystore"
	"github.com/centrifuge/go-centrifuge/pallets/proxy"
	"github.com/centrifuge/go-centrifuge/storage/leveldb"
	testingcommons "github.com/centrifuge/go-centrifuge/testingutils/common"
	"github.com/centrifuge/go-centrifuge/testingutils/keyrings"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/stretchr/testify/assert"
)

var integrationTestBootstrappers = []bootstrap.TestBootstrapper{
	&testlogging.TestLoggingBootstrapper{},
	&config.Bootstrapper{},
	&leveldb.Bootstrapper{},
	&configstore.Bootstrapper{},
	&jobs.Bootstrapper{},
	&integration_test.Bootstrapper{},
	centchain.Bootstrapper{},
	&pallets.Bootstrapper{},
	&dispatcher.Bootstrapper{},
	&v2.Bootstrapper{},
}

var (
	cfgService  config.Service
	centAPI     centchain.API
	keystoreAPI keystore.API
	proxyAPI    proxy.API
)

func TestMain(m *testing.M) {
	ctx := bootstrap.RunTestBootstrappers(integrationTestBootstrappers, nil)
	cfgService = ctx[config.BootstrappedConfigStorage].(config.Service)
	centAPI = ctx[centchain.BootstrappedCentChainClient].(centchain.API)
	keystoreAPI = ctx[pallets.BootstrappedKeystoreAPI].(keystore.API)
	proxyAPI = ctx[pallets.BootstrappedProxyAPI].(proxy.API)

	result := m.Run()

	bootstrap.RunTestTeardown(integrationTestBootstrappers)

	os.Exit(result)
}

func TestIntegration_API_ProxyCall(t *testing.T) {
	// There is an account created for Alice by the identity test bootstrapper.
	// We are going to use that account to test that the pod operator can do a call on
	// Alice's behalf.

	ctx := context.Background()

	acc, err := cfgService.GetAccount(keyrings.AliceKeyRingPair.PublicKey)
	assert.NoError(t, err)

	ctx = contextutil.WithAccount(ctx, acc)

	podOperator, err := cfgService.GetPodOperator()
	assert.NoError(t, err)

	forcedProxyType := types.NewOption[proxyTypes.CentrifugeProxyType](proxyTypes.KeystoreManagement)

	meta, err := centAPI.GetMetadataLatest()
	assert.NoError(t, err)

	keyHash := types.NewHash(utils.RandomSlice(32))

	keys := []*keystoreType.AddKey{
		{
			Key:     keyHash,
			Purpose: keystoreType.KeyPurposeP2PDiscovery,
			KeyType: keystoreType.KeyTypeECDSA,
		},
	}

	proxiedCall, err := types.NewCall(meta, keystore.AddKeysCall, keys)
	assert.NoError(t, err)

	extInfo, err := proxyAPI.ProxyCall(ctx, acc.GetIdentity(), podOperator.ToKeyringPair(), forcedProxyType, proxiedCall)
	assert.NoError(t, err)
	assert.NotNil(t, extInfo)

	// Confirm that the key was added.

	_, err = keystoreAPI.GetKey(acc.GetIdentity(), &keystoreType.KeyID{
		Hash:       keyHash,
		KeyPurpose: keystoreType.KeyPurposeP2PDiscovery,
	})
	assert.NoError(t, err)
}

func TestIntegration_API_AddAndRetrieveProxies(t *testing.T) {
	ctx := context.Background()

	delegate1, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	delegate1ProxyType := proxyTypes.Borrow

	delegate2, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	delegate2ProxyType := proxyTypes.Invest

	// We are using Alice's account to add the proxies since it has the necessary balance for it.

	err = proxyAPI.AddProxy(ctx, delegate1, delegate1ProxyType, 0, keyrings.AliceKeyRingPair)
	assert.NoError(t, err)

	err = proxyAPI.AddProxy(ctx, delegate2, delegate2ProxyType, 0, keyrings.AliceKeyRingPair)
	assert.NoError(t, err)

	accountIDAlice, err := types.NewAccountID(keyrings.AliceKeyRingPair.PublicKey)
	assert.NoError(t, err)

	err = waitForProxiesToBeAdded(ctx, accountIDAlice, delegate1, delegate2)
	assert.NoError(t, err)
}

const (
	proxiesCheckInterval = 5 * time.Second
	proxiesCheckTimeout  = 1 * time.Minute
)

func waitForProxiesToBeAdded(
	ctx context.Context,
	delegatorAccountID *types.AccountID,
	expectedProxies ...*types.AccountID,
) error {
	expectedProxiesMap := make(map[string]struct{})

	for _, expectedProxy := range expectedProxies {
		expectedProxiesMap[expectedProxy.ToHexString()] = struct{}{}
	}

	ctx, cancel := context.WithTimeout(ctx, proxiesCheckTimeout)
	defer cancel()

	t := time.NewTicker(proxiesCheckInterval)
	defer t.Stop()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("context expired while checking for proxies: %w", ctx.Err())
		case <-t.C:
			res, err := proxyAPI.GetProxies(delegatorAccountID)

			if err != nil {
				continue
			}

			for _, proxyDefinition := range res.ProxyDefinitions {
				delete(expectedProxiesMap, proxyDefinition.Delegate.ToHexString())
			}

			if len(expectedProxiesMap) == 0 {
				return nil
			}
		}
	}
}
