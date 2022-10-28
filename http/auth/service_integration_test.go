//go:build integration

package auth

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/centrifuge/go-substrate-rpc-client/v4/signature"

	proxyType "github.com/centrifuge/chain-custom-types/pkg/proxy"
	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/bootstrap/bootstrappers/integration_test"
	"github.com/centrifuge/go-centrifuge/bootstrap/bootstrappers/testlogging"
	"github.com/centrifuge/go-centrifuge/centchain"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/config/configstore"
	protocolIDDispatcher "github.com/centrifuge/go-centrifuge/dispatcher"
	identityV2 "github.com/centrifuge/go-centrifuge/identity/v2"
	"github.com/centrifuge/go-centrifuge/jobs"
	"github.com/centrifuge/go-centrifuge/pallets"
	"github.com/centrifuge/go-centrifuge/pallets/proxy"
	"github.com/centrifuge/go-centrifuge/storage/leveldb"
	"github.com/centrifuge/go-centrifuge/testingutils/keyrings"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/stretchr/testify/assert"
	"github.com/vedhavyas/go-subkey/v2"
	"github.com/vedhavyas/go-subkey/v2/sr25519"
)

var integrationTestBootstrappers = []bootstrap.TestBootstrapper{
	&testlogging.TestLoggingBootstrapper{},
	&config.Bootstrapper{},
	&leveldb.Bootstrapper{},
	&jobs.Bootstrapper{},
	&configstore.Bootstrapper{},
	&integration_test.Bootstrapper{},
	centchain.Bootstrapper{},
	&pallets.Bootstrapper{},
	&protocolIDDispatcher.Bootstrapper{},
	&identityV2.AccountTestBootstrapper{},
}

var (
	proxyAPI  proxy.API
	configSrv config.Service
)

func TestMain(m *testing.M) {
	ctx := bootstrap.RunTestBootstrappers(integrationTestBootstrappers, nil)
	proxyAPI = ctx[pallets.BootstrappedProxyAPI].(proxy.API)
	configSrv = ctx[config.BootstrappedConfigStorage].(config.Service)

	// Add Bob as PodAuth proxy to Alice.
	if err := setupPodAuthProxy(keyrings.AliceKeyRingPair, keyrings.BobKeyRingPair.PublicKey); err != nil {
		panic(err)
	}

	result := m.Run()

	bootstrap.RunTestTeardown(integrationTestBootstrappers)

	os.Exit(result)
}

func setupPodAuthProxy(delegatorKeyringPair signature.KeyringPair, delegatePublicKey []byte) error {
	delegateAccountID, err := types.NewAccountID(delegatePublicKey)
	if err != nil {
		return fmt.Errorf("couldn't create delegate account ID: %w", err)
	}

	err = proxyAPI.AddProxy(context.Background(), delegateAccountID, proxyType.PodAuth, 0, delegatorKeyringPair)

	if err != nil {
		return fmt.Errorf("couldn't add pod auth proxy: %w", err)
	}

	return nil
}

func TestIntegration_Service_Validate(t *testing.T) {
	srv := NewService(true, proxyAPI, configSrv)

	delegateAccountID, err := types.NewAccountID(keyrings.BobKeyRingPair.PublicKey)
	assert.NoError(t, err)

	delegatorAccountID, err := types.NewAccountID(keyrings.AliceKeyRingPair.PublicKey)
	assert.NoError(t, err)

	token, err := CreateJW3Token(
		delegateAccountID,
		delegatorAccountID,
		keyrings.BobKeyRingPair.URI,
		proxyType.ProxyTypeName[proxyType.PodAuth],
	)
	assert.NoError(t, err)

	ctx := context.Background()

	res, err := srv.Validate(ctx, token)
	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.False(t, res.IsAdmin)
}

func TestIntegration_Service_Validate_ProxyTypeMismatch(t *testing.T) {
	srv := NewService(true, proxyAPI, configSrv)

	delegateAccountID, err := types.NewAccountID(keyrings.BobKeyRingPair.PublicKey)
	assert.NoError(t, err)

	delegatorAccountID, err := types.NewAccountID(keyrings.AliceKeyRingPair.PublicKey)
	assert.NoError(t, err)

	token, err := CreateJW3Token(
		delegateAccountID,
		delegatorAccountID,
		keyrings.BobKeyRingPair.URI,
		// Bob is added as proxy to Alice only as PodAuth.
		proxyType.ProxyTypeName[proxyType.PodOperation],
	)
	assert.NoError(t, err)

	ctx := context.Background()

	res, err := srv.Validate(ctx, token)
	assert.ErrorIs(t, err, ErrInvalidDelegate)
	assert.Nil(t, res)
}

func TestIntegration_Service_Validate_InvalidIdentity(t *testing.T) {
	srv := NewService(true, proxyAPI, configSrv)

	delegateAccountID, err := types.NewAccountID(keyrings.BobKeyRingPair.PublicKey)
	assert.NoError(t, err)

	// There is no identity created for Charlie.
	delegatorAccountID, err := types.NewAccountID(keyrings.CharlieKeyRingPair.PublicKey)
	assert.NoError(t, err)

	token, err := CreateJW3Token(
		delegateAccountID,
		delegatorAccountID,
		keyrings.BobKeyRingPair.URI,
		proxyType.ProxyTypeName[proxyType.PodAuth],
	)
	assert.NoError(t, err)

	ctx := context.Background()

	res, err := srv.Validate(ctx, token)
	assert.ErrorIs(t, err, ErrInvalidIdentity)
	assert.Nil(t, res)
}

func TestIntegration_Service_Validate_PodAdmin(t *testing.T) {
	srv := NewService(true, proxyAPI, configSrv)

	cfg, err := configSrv.GetConfig()
	assert.NoError(t, err)

	podAdminKeyPair, err := subkey.DeriveKeyPair(sr25519.Scheme{}, cfg.GetPodAdminSecretSeed())
	assert.NoError(t, err)

	podAdminAccountID, err := types.NewAccountID(podAdminKeyPair.AccountID())
	assert.NoError(t, err)

	token, err := CreateJW3Token(
		podAdminAccountID,
		podAdminAccountID,
		hexutil.Encode(podAdminKeyPair.Seed()),
		PodAdminProxyType,
	)
	assert.NoError(t, err)

	ctx := context.Background()

	res, err := srv.Validate(ctx, token)
	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.True(t, res.IsAdmin)
}

func TestIntegration_Service_Validate_PodAdmin_Error(t *testing.T) {
	srv := NewService(true, proxyAPI, configSrv)

	podAdmin, err := configSrv.GetPodAdmin()
	assert.NoError(t, err)

	randomKeyPair, err := sr25519.Scheme{}.Generate()
	assert.NoError(t, err)

	randomAccountID, err := types.NewAccountID(randomKeyPair.AccountID())
	assert.NoError(t, err)

	token, err := CreateJW3Token(
		randomAccountID,
		podAdmin.GetAccountID(),
		hexutil.Encode(randomKeyPair.Seed()),
		PodAdminProxyType,
	)
	assert.NoError(t, err)

	ctx := context.Background()

	res, err := srv.Validate(ctx, token)
	assert.ErrorIs(t, err, ErrNotAdminAccount)
	assert.Nil(t, res)
}
