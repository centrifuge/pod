// +build unit

package userapi

import (
	"os"
	"testing"

	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/bootstrap/bootstrappers/testlogging"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/config/configstore"
	"github.com/centrifuge/go-centrifuge/documents/entity"
	"github.com/centrifuge/go-centrifuge/documents/entityrelationship"
	"github.com/centrifuge/go-centrifuge/extensions/funding"
	"github.com/centrifuge/go-centrifuge/extensions/transferdetails"
	"github.com/centrifuge/go-centrifuge/httpapi/coreapi"
	"github.com/centrifuge/go-centrifuge/jobs/jobsv1"
	"github.com/centrifuge/go-centrifuge/storage/leveldb"
	testingidentity "github.com/centrifuge/go-centrifuge/testingutils/identity"
	"github.com/stretchr/testify/assert"
)

var ctx = map[string]interface{}{}
var cfg config.Configuration
var did = testingidentity.GenerateRandomDID()

func TestMain(m *testing.M) {
	ibootstappers := []bootstrap.TestBootstrapper{
		&testlogging.TestLoggingBootstrapper{},
		&config.Bootstrapper{},
		&leveldb.Bootstrapper{},
		jobsv1.Bootstrapper{},
	}
	bootstrap.RunTestBootstrappers(ibootstappers, ctx)
	cfg = ctx[bootstrap.BootstrappedConfig].(config.Configuration)
	cfg.Set("keys.p2p.publicKey", "../../build/resources/p2pKey.pub.pem")
	cfg.Set("keys.p2p.privateKey", "../../build/resources/p2pKey.key.pem")
	cfg.Set("keys.signing.publicKey", "../../build/resources/signingKey.pub.pem")
	cfg.Set("keys.signing.privateKey", "../../build/resources/signingKey.key.pem")
	cfg.Set("networks.testing.contractAddresses.invoiceUnpaid", "0xf72855759a39fb75fc7341139f5d7a3974d4da08")
	cfg.Set("identityId", did.String())
	result := m.Run()
	bootstrap.RunTestTeardown(ibootstappers)
	os.Exit(result)
}

func TestBootstrapper_Bootstrap(t *testing.T) {
	ctx := make(map[string]interface{})
	b := Bootstrapper{}

	// missing core-api service
	err := b.Bootstrap(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), coreapi.BootstrappedCoreAPIService)

	// missing transfer detail service
	ctx[coreapi.BootstrappedCoreAPIService] = coreapi.Service{}
	err = b.Bootstrap(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), transferdetails.BootstrappedTransferDetailService)

	// missing entityrelationship service
	ctx[transferdetails.BootstrappedTransferDetailService] = new(MockTransferService)
	err = b.Bootstrap(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), entityrelationship.BootstrappedEntityRelationshipService)

	// missing entity service
	ctx[entityrelationship.BootstrappedEntityRelationshipService] = new(entity.MockEntityRelationService)
	err = b.Bootstrap(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), entity.BootstrappedEntityService)

	// missing funding service
	ctx[entity.BootstrappedEntityService] = new(entity.MockService)
	err = b.Bootstrap(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), funding.BootstrappedFundingService)

	// missing config service
	ctx[funding.BootstrappedFundingService] = new(funding.MockService)
	err = b.Bootstrap(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), config.BootstrappedConfigStorage)

	// success
	ctx[config.BootstrappedConfigStorage] = new(configstore.MockService)
	assert.NoError(t, b.Bootstrap(ctx))
}
