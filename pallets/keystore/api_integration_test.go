//go:build integration

package keystore_test

import (
	"context"
	"os"
	"testing"

	keystoreType "github.com/centrifuge/chain-custom-types/pkg/keystore"
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
	"github.com/centrifuge/go-centrifuge/pallets"
	"github.com/centrifuge/go-centrifuge/pallets/keystore"
	"github.com/centrifuge/go-centrifuge/storage/leveldb"
	genericUtils "github.com/centrifuge/go-centrifuge/testingutils/generic"
	"github.com/centrifuge/go-centrifuge/utils"
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

var (
	cfgService  config.Service
	keystoreAPI keystore.API
)

func TestMain(m *testing.M) {
	ctx := bootstrap.RunTestBootstrappers(integrationTestBootstrappers, nil)
	cfgService = genericUtils.GetService[config.Service](ctx)
	keystoreAPI = genericUtils.GetService[keystore.API](ctx)

	result := m.Run()

	bootstrap.RunTestTeardown(integrationTestBootstrappers)

	os.Exit(result)
}

func TestIntegration_API_KeyOperations(t *testing.T) {
	accs, err := cfgService.GetAccounts()
	assert.NoError(t, err)
	assert.NotEmpty(t, accs)

	acc := accs[0]

	ctx := contextutil.WithAccount(context.Background(), acc)

	// Add keys

	key1Hash := types.NewHash(utils.RandomSlice(32))
	key1Purpose := keystoreType.KeyPurposeP2PDiscovery
	key1Type := keystoreType.KeyTypeECDSA

	key2Hash := types.NewHash(utils.RandomSlice(32))
	key2Purpose := keystoreType.KeyPurposeP2PDocumentSigning
	key2Type := keystoreType.KeyTypeEDDSA

	keys := []*keystoreType.AddKey{
		{
			Key:     key1Hash,
			Purpose: key1Purpose,
			KeyType: key1Type,
		},
		{
			Key:     key2Hash,
			Purpose: key2Purpose,
			KeyType: key2Type,
		},
	}

	extInfo, err := keystoreAPI.AddKeys(ctx, keys)
	assert.NoError(t, err)
	assert.NotNil(t, extInfo)

	// Retrieve keys

	key1, err := keystoreAPI.GetKey(acc.GetIdentity(), &keystoreType.KeyID{
		Hash:       key1Hash,
		KeyPurpose: key1Purpose,
	})
	assert.NoError(t, err)

	assert.Equal(t, key1Type, key1.KeyType)
	assert.Equal(t, key1Purpose, key1.KeyPurpose)

	key1, err = keystoreAPI.GetKey(acc.GetIdentity(), &keystoreType.KeyID{
		Hash:       key1Hash,
		KeyPurpose: key2Purpose,
	})
	assert.ErrorIs(t, err, keystore.ErrKeyNotFound)

	// TODO(cdamian) Re-enable when we have tests that are using anonymous proxies for identities.
	//keyHash, err := keystoreAPI.GetLastKeyByPurpose(acc.GetIdentity(), key1Purpose)
	//assert.NoError(t, err)
	//assert.Equal(t, key1Hash, *keyHash)

	key2, err := keystoreAPI.GetKey(acc.GetIdentity(), &keystoreType.KeyID{
		Hash:       key2Hash,
		KeyPurpose: key2Purpose,
	})
	assert.NoError(t, err)

	assert.Equal(t, key2Type, key2.KeyType)
	assert.Equal(t, key2Purpose, key2.KeyPurpose)

	key2, err = keystoreAPI.GetKey(acc.GetIdentity(), &keystoreType.KeyID{
		Hash:       key2Hash,
		KeyPurpose: key1Purpose,
	})
	assert.ErrorIs(t, err, keystore.ErrKeyNotFound)

	// TODO(cdamian) Re-enable when we have tests that are using anonymous proxies for identities.
	//keyHash, err = keystoreAPI.GetLastKeyByPurpose(acc.GetIdentity(), key2Purpose)
	//assert.NoError(t, err)
	//assert.Equal(t, key2Hash, *keyHash)

	// Revoke keys

	extInfo, err = keystoreAPI.RevokeKeys(ctx, []*types.Hash{&key1Hash}, key1Purpose)
	assert.NoError(t, err)
	assert.NotNil(t, extInfo)

	extInfo, err = keystoreAPI.RevokeKeys(ctx, []*types.Hash{&key2Hash}, key2Purpose)
	assert.NoError(t, err)
	assert.NotNil(t, extInfo)

	// Confirm that keys were revoked

	key1, err = keystoreAPI.GetKey(acc.GetIdentity(), &keystoreType.KeyID{
		Hash:       key1Hash,
		KeyPurpose: key1Purpose,
	})
	assert.NoError(t, err)
	assert.True(t, key1.RevokedAt.HasValue())

	key2, err = keystoreAPI.GetKey(acc.GetIdentity(), &keystoreType.KeyID{
		Hash:       key2Hash,
		KeyPurpose: key2Purpose,
	})
	assert.NoError(t, err)
	assert.True(t, key2.RevokedAt.HasValue())
}
