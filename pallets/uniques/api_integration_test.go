//go:build integration

package uniques_test

import (
	"context"
	"math/big"
	"math/rand"
	"os"
	"testing"
	"time"

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
	"github.com/centrifuge/go-centrifuge/pallets/uniques"
	"github.com/centrifuge/go-centrifuge/storage/leveldb"
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
	&v2.AccountTestBootstrapper{},
}

var (
	cfgService config.Service
	uniquesAPI uniques.API
)

func TestMain(m *testing.M) {
	ctx := bootstrap.RunTestBootstrappers(integrationTestBootstrappers, nil)
	cfgService = ctx[config.BootstrappedConfigStorage].(config.Service)
	uniquesAPI = ctx[pallets.BootstrappedUniquesAPI].(uniques.API)

	rand.Seed(time.Now().Unix())

	result := m.Run()

	bootstrap.RunTestTeardown(integrationTestBootstrappers)

	os.Exit(result)
}

func TestIntegration_API_NFTOperations(t *testing.T) {
	acc, err := cfgService.GetAccount(keyrings.AliceKeyRingPair.PublicKey)
	assert.NoError(t, err)

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
