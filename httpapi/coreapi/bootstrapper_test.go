// +build unit

package coreapi

import (
	"testing"

	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/config/configstore"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/jobs"
	testingdocuments "github.com/centrifuge/go-centrifuge/testingutils/documents"
	testingnfts "github.com/centrifuge/go-centrifuge/testingutils/nfts"
	"github.com/centrifuge/go-centrifuge/testingutils/testingjobs"
	"github.com/stretchr/testify/assert"
)

func TestBootstrapper_Bootstrap(t *testing.T) {
	ctx := make(map[string]interface{})
	b := Bootstrapper{}

	// missing doc service
	err := b.Bootstrap(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), documents.BootstrappedDocumentService)

	// missing jobs service
	ctx[documents.BootstrappedDocumentService] = new(testingdocuments.MockService)
	err = b.Bootstrap(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), jobs.BootstrappedService)

	// missing nft service
	ctx[jobs.BootstrappedService] = new(testingjobs.MockJobManager)
	err = b.Bootstrap(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), bootstrap.BootstrappedInvoiceUnpaid)

	// missing accounts service
	ctx[bootstrap.BootstrappedInvoiceUnpaid] = new(testingnfts.MockNFTService)
	err = b.Bootstrap(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), config.BootstrappedConfigStorage)

	// success
	ctx[config.BootstrappedConfigStorage] = new(configstore.MockService)
	assert.NoError(t, b.Bootstrap(ctx))
}
