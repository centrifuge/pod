// +build unit

package v2

import (
	"testing"

	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/pending"
	testingnfts "github.com/centrifuge/go-centrifuge/testingutils/nfts"
	"github.com/stretchr/testify/assert"
)

func TestBootstrapper_Bootstrap(t *testing.T) {
	ctx := make(map[string]interface{})
	b := Bootstrapper{}

	// missing pending doc service
	err := b.Bootstrap(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), pending.BootstrappedPendingDocumentService)

	// missing nft service
	ctx[pending.BootstrappedPendingDocumentService] = new(pending.MockService)
	err = b.Bootstrap(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), bootstrap.BootstrappedInvoiceUnpaid)

	// success
	ctx[bootstrap.BootstrappedInvoiceUnpaid] = new(testingnfts.MockNFTService)
	err = b.Bootstrap(ctx)
	assert.NoError(t, b.Bootstrap(ctx))
	assert.NotNil(t, ctx[BootstrappedService])
}
