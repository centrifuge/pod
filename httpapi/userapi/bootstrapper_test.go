// +build unit

package userapi

import (
	"testing"

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

	// success
	ctx[transferdetails.BootstrappedTransferDetailService] = new(MockTransferService)
	assert.NoError(t, b.Bootstrap(ctx))
}
