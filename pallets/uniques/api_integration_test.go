//go:build integration

package uniques_test

import (
	"context"
	"math/big"
	"math/rand"
	"os"
	"testing"
	"time"

	genericUtils "github.com/centrifuge/pod/testingutils/generic"

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
	"github.com/centrifuge/pod/storage/leveldb"
	"github.com/centrifuge/pod/utils"
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
)

func TestMain(m *testing.M) {
	ctx := bootstrap.RunTestBootstrappers(integrationTestBootstrappers, nil)
	cfgService = genericUtils.GetService[config.Service](ctx)
	uniquesAPI = genericUtils.GetService[uniques.API](ctx)

	rand.Seed(time.Now().Unix())

	result := m.Run()

	bootstrap.RunTestTeardown(integrationTestBootstrappers)

	os.Exit(result)
}

func TestIntegration_API_NFTOperations(t *testing.T) {
	accs, err := cfgService.GetAccounts()
	assert.NoError(t, err)
	assert.NotEmpty(t, accs)

	acc := accs[0]

	ctx := contextutil.WithAccount(context.Background(), acc)

	collectionID := types.U64(rand.Uint64())
	itemID := types.NewU128(*big.NewInt(rand.Int63()))

	itemMetadataData := utils.RandomSlice(32)
	itemAttributeKey := utils.RandomSlice(32)
	itemAttributeValue := utils.RandomSlice(32)

	// Create NFT collection.

	_, err = uniquesAPI.CreateCollection(ctx, collectionID)
	assert.NoError(t, err)

	// Mint NFT.

	_, err = uniquesAPI.Mint(ctx, collectionID, itemID, acc.GetIdentity())
	assert.NoError(t, err)

	collectionDetails, err := uniquesAPI.GetCollectionDetails(collectionID)
	assert.NoError(t, err)
	assert.True(t, collectionDetails.Owner.Equal(acc.GetIdentity()))

	itemDetails, err := uniquesAPI.GetItemDetails(collectionID, itemID)
	assert.NoError(t, err)
	assert.True(t, itemDetails.Owner.Equal(acc.GetIdentity()))

	// Set NFT metadata.

	_, err = uniquesAPI.SetMetadata(ctx, collectionID, itemID, itemMetadataData, false)
	assert.NoError(t, err)

	itemMetadata, err := uniquesAPI.GetItemMetadata(collectionID, itemID)
	assert.NoError(t, err)
	assert.Equal(t, []byte(itemMetadata.Data), itemMetadataData)
	assert.False(t, itemMetadata.IsFrozen)

	// Set NFT attributes.

	_, err = uniquesAPI.SetAttribute(ctx, collectionID, itemID, itemAttributeKey, itemAttributeValue)
	assert.NoError(t, err)

	itemAttribute, err := uniquesAPI.GetItemAttribute(collectionID, itemID, itemAttributeKey)
	assert.NoError(t, err)
	assert.Equal(t, itemAttributeValue, itemAttribute)
}
