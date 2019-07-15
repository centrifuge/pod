// +build unit

package userapi

import (
	"testing"

	"github.com/centrifuge/go-centrifuge/documents/entity"
	"github.com/centrifuge/go-centrifuge/documents/entityrelationship"
	"github.com/centrifuge/go-centrifuge/extensions/funding"
	"github.com/centrifuge/go-centrifuge/extensions/transferdetails"
	"github.com/centrifuge/go-centrifuge/httpapi/coreapi"
	"github.com/stretchr/testify/assert"
)

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

	// success
	ctx[funding.BootstrappedFundingService] = new(funding.MockService)
	assert.NoError(t, b.Bootstrap(ctx))
}
