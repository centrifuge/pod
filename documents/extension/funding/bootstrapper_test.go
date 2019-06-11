// +build unit

package funding

import (
	"testing"

	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/config/configstore"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/testingutils/documents"
	"github.com/stretchr/testify/assert"
)

func TestBootstrapper_Bootstrap(t *testing.T) {
	ctx := make(map[string]interface{})
	b := Bootstrapper{}

	// missing config service
	err := b.Bootstrap(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "funding bootstrapper: config service not initialised")

	// missing doc service
	ctx[config.BootstrappedConfigStorage] = new(configstore.MockService)
	err = b.Bootstrap(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "document service not initialised")

	// missing identity service
	ctx[documents.BootstrappedDocumentService] = new(mockService)
	err = b.Bootstrap(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "token registry not initialisation")

	// success
	ctx[bootstrap.BootstrappedInvoiceUnpaid] = new(testingdocuments.MockRegistry)
	err = b.Bootstrap(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, ctx[BootstrappedFundingAPIHandler])
}
