//go:build integration

package utility_test

import (
	"context"
	"math/rand"
	"os"
	"testing"
	"time"

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
	"github.com/centrifuge/pod/pallets/uniques"
	"github.com/centrifuge/pod/pallets/utility"
	"github.com/centrifuge/pod/storage/leveldb"
	genericUtils "github.com/centrifuge/pod/testingutils/generic"
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
	cfgService config.Service
	uniquesAPI uniques.API
	utilityAPI utility.API
)

func TestMain(m *testing.M) {
	ctx := bootstrap.RunTestBootstrappers(integrationTestBootstrappers, nil)
	cfgService = genericUtils.GetService[config.Service](ctx)
	uniquesAPI = genericUtils.GetService[uniques.API](ctx)
	utilityAPI = genericUtils.GetService[utility.API](ctx)

	rand.Seed(time.Now().Unix())

	result := m.Run()

	bootstrap.RunTestTeardown(integrationTestBootstrappers)

	os.Exit(result)
}

func TestIntegration_API_BatchAll(t *testing.T) {
	accs, err := cfgService.GetAccounts()
	assert.NoError(t, err)
	assert.NotEmpty(t, accs)

	acc := accs[0]

	ctx := contextutil.WithAccount(context.Background(), acc)

	collectionID1 := types.U64(rand.Uint64())
	collectionID2 := types.U64(rand.Uint64())

	collectionAdminMultiAddress, err := types.NewMultiAddressFromAccountID(acc.GetIdentity().ToBytes())
	assert.NoError(t, err)

	callCreationFn1 := centchain.CallProviderFn(func(meta *types.Metadata) (*types.Call, error) {
		call, err := types.NewCall(
			meta,
			uniques.CreateCollectionCall,
			collectionID1,
			collectionAdminMultiAddress,
		)

		if err != nil {
			return nil, err
		}

		return &call, nil
	})

	callCreationFn2 := centchain.CallProviderFn(func(meta *types.Metadata) (*types.Call, error) {
		call, err := types.NewCall(
			meta,
			uniques.CreateCollectionCall,
			collectionID2,
			collectionAdminMultiAddress,
		)

		if err != nil {
			return nil, err
		}

		return &call, nil
	})

	_, err = utilityAPI.BatchAll(ctx, callCreationFn1, callCreationFn2)
	assert.NoError(t, err)

	collectionDetails, err := uniquesAPI.GetCollectionDetails(collectionID1)
	assert.NoError(t, err)
	assert.True(t, collectionDetails.Owner.Equal(acc.GetIdentity()))

	collectionDetails, err = uniquesAPI.GetCollectionDetails(collectionID2)
	assert.NoError(t, err)
	assert.True(t, collectionDetails.Owner.Equal(acc.GetIdentity()))
}
