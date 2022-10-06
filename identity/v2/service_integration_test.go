//go:build integration

package v2

import (
	"context"
	"os"
	"testing"
	"time"

	proxyType "github.com/centrifuge/chain-custom-types/pkg/proxy"

	"github.com/centrifuge/go-centrifuge/pallets/proxy"

	testingcommons "github.com/centrifuge/go-centrifuge/testingutils/common"

	"github.com/centrifuge/go-centrifuge/contextutil"

	"github.com/centrifuge/go-centrifuge/utils"

	keystoreType "github.com/centrifuge/chain-custom-types/pkg/keystore"

	"github.com/centrifuge/go-centrifuge/testingutils/keyrings"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"

	"github.com/stretchr/testify/assert"

	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/bootstrap/bootstrappers/integration_test"
	"github.com/centrifuge/go-centrifuge/bootstrap/bootstrappers/testlogging"
	"github.com/centrifuge/go-centrifuge/centchain"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/config/configstore"
	protocolIDDispatcher "github.com/centrifuge/go-centrifuge/dispatcher"
	"github.com/centrifuge/go-centrifuge/jobs"
	"github.com/centrifuge/go-centrifuge/pallets"
	"github.com/centrifuge/go-centrifuge/pallets/keystore"
	"github.com/centrifuge/go-centrifuge/storage/leveldb"
)

var integrationTestBootstrappers = []bootstrap.TestBootstrapper{
	&testlogging.TestLoggingBootstrapper{},
	&config.Bootstrapper{},
	&leveldb.Bootstrapper{},
	jobs.Bootstrapper{},
	&configstore.Bootstrapper{},
	&integration_test.Bootstrapper{},
	centchain.Bootstrapper{},
	&pallets.Bootstrapper{},
	&protocolIDDispatcher.Bootstrapper{},
	&Bootstrapper{},
}

var (
	configSrv       config.Service
	centAPI         centchain.API
	keystoreAPI     keystore.API
	proxyAPI        proxy.API
	identityService Service
)

func TestMain(m *testing.M) {
	ctx := bootstrap.RunTestBootstrappers(integrationTestBootstrappers, nil)
	configSrv = ctx[config.BootstrappedConfigStorage].(config.Service)
	centAPI = ctx[centchain.BootstrappedCentChainClient].(centchain.API)
	keystoreAPI = ctx[pallets.BootstrappedKeystoreAPI].(keystore.API)
	identityService = ctx[BootstrappedIdentityServiceV2].(Service)
	proxyAPI = ctx[pallets.BootstrappedProxyAPI].(proxy.API)

	result := m.Run()

	bootstrap.RunTestTeardown(integrationTestBootstrappers)

	os.Exit(result)
}

func TestIntegration_Service_CreateIdentity(t *testing.T) {
	ctx := context.Background()

	// Dave has an account on chain.
	accountID, err := types.NewAccountID(keyrings.DaveKeyRingPair.PublicKey)
	assert.NoError(t, err)

	defer func() {
		// Ensure that we clean up.
		_ = configSrv.DeleteAccount(accountID.ToBytes())
	}()

	req := &CreateIdentityRequest{
		Identity:         accountID,
		WebhookURL:       "https://centrifuge.io",
		PrecommitEnabled: true,
	}

	acc, err := identityService.CreateIdentity(ctx, req)
	assert.NoError(t, err)
	assert.NotNil(t, acc)

	// Identity already exists, should error out during account storage.
	acc, err = identityService.CreateIdentity(ctx, req)
	assert.NotNil(t, err)
	assert.Nil(t, acc)
}

func TestIntegration_Service_ValidateKey(t *testing.T) {
	ctx := context.Background()

	// We are using Alice's account since the pod operator is added as a proxy.
	acc, err := configSrv.GetAccount(keyrings.AliceKeyRingPair.PublicKey)
	assert.NoError(t, err)

	ctx = contextutil.WithAccount(ctx, acc)

	testKey := utils.RandomSlice(32)

	_, err = keystoreAPI.AddKeys(ctx, []*keystoreType.AddKey{
		{
			Key:     types.NewHash(testKey),
			Purpose: keystoreType.KeyPurposeP2PDocumentSigning,
			KeyType: keystoreType.KeyTypeECDSA,
		},
	})
	assert.NoError(t, err)

	validationTime := time.Now()

	err = identityService.ValidateKey(ctx, acc.GetIdentity(), testKey, keystoreType.KeyPurposeP2PDocumentSigning, validationTime)
	assert.NoError(t, err)

	keyHash := types.NewHash(testKey)

	_, err = keystoreAPI.RevokeKeys(ctx, []*types.Hash{&keyHash}, keystoreType.KeyPurposeP2PDocumentSigning)
	assert.NoError(t, err)

	validationTime = time.Now()

	err = identityService.ValidateKey(ctx, acc.GetIdentity(), testKey, keystoreType.KeyPurposeP2PDocumentSigning, validationTime)
	assert.ErrorIs(t, err, ErrKeyRevoked)

	validationTime = validationTime.Add(-1 * time.Hour)

	err = identityService.ValidateKey(ctx, acc.GetIdentity(), testKey, keystoreType.KeyPurposeP2PDocumentSigning, validationTime)
	assert.NoError(t, err)
}

func TestIntegration_Service_ValidateSignature(t *testing.T) {
	ctx := context.Background()

	acc, err := configSrv.GetAccount(keyrings.AliceKeyRingPair.PublicKey)
	assert.NoError(t, err)

	// This key pair is added by the test bootstrapper.
	_, signingPrivateKey, err := testingcommons.GetTestSigningKeys()
	assert.NoError(t, err)

	message := utils.RandomSlice(32)

	signature, err := signingPrivateKey.Sign(message)
	assert.NoError(t, err)

	err = identityService.ValidateSignature(ctx, acc.GetIdentity(), acc.GetSigningPublicKey(), message, signature, time.Now())
	assert.NoError(t, err)

	signature = utils.RandomSlice(32)

	err = identityService.ValidateSignature(ctx, acc.GetIdentity(), acc.GetSigningPublicKey(), message, signature, time.Now())
	assert.ErrorIs(t, err, ErrInvalidSignature)
}

func TestIntegration_Service_ValidateAccount(t *testing.T) {
	ctx := context.Background()

	devAccountPubKeys := [][]byte{
		keyrings.AliceKeyRingPair.PublicKey,
		keyrings.BobKeyRingPair.PublicKey,
		keyrings.CharlieKeyRingPair.PublicKey,
		keyrings.DaveKeyRingPair.PublicKey,
		keyrings.EveKeyRingPair.PublicKey,
		keyrings.FerdieKeyRingPair.PublicKey,
	}

	for _, devAccountPubKey := range devAccountPubKeys {
		accID, err := types.NewAccountID(devAccountPubKey)
		assert.NoError(t, err)

		err = identityService.ValidateAccount(ctx, accID)
		assert.NoError(t, err)
	}

	randomProxiedAccountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	// Add a proxy with type any to a random account.
	err = proxyAPI.AddProxy(ctx, randomProxiedAccountID, proxyType.Any, 0, keyrings.AliceKeyRingPair)
	assert.NoError(t, err)

	err = identityService.ValidateAccount(ctx, randomProxiedAccountID)
	assert.ErrorIs(t, err, ErrAccountNotAnonymousProxy)

	// Test with a random account ID that has no account info nor a proxy with type any.
	randomAccountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	err = identityService.ValidateAccount(ctx, randomAccountID)
	assert.ErrorIs(t, err, ErrAccountNotAnonymousProxy)

	err = proxyAPI.AddProxy(ctx, randomProxiedAccountID, proxyType.PodAuth, 0, keyrings.AliceKeyRingPair)
	assert.NoError(t, err)

	err = identityService.ValidateAccount(ctx, randomAccountID)
	assert.ErrorIs(t, err, ErrAccountNotAnonymousProxy)
}

func TestIntegration_Service_GetLastKeyByPurpose(t *testing.T) {
	ctx := context.Background()

	// Alice has a p2p document signing key added by the test bootstrapper.
	accountID, err := types.NewAccountID(keyrings.AliceKeyRingPair.PublicKey)
	assert.NoError(t, err)

	key, err := identityService.GetLastKeyByPurpose(ctx, accountID, keystoreType.KeyPurposeP2PDocumentSigning)
	assert.NoError(t, err)
	assert.NotNil(t, key)

	key, err = identityService.GetLastKeyByPurpose(ctx, accountID, keystoreType.KeyPurposeP2PDiscovery)
	assert.ErrorIs(t, err, ErrKeyRetrieval)
	assert.Nil(t, key)
}
