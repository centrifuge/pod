// +build unit

package transferdetails

import (
	"testing"

	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/extensions"
	"github.com/centrifuge/go-centrifuge/testingutils/documents"
	"github.com/stretchr/testify/assert"
)

func TestBootstrapper_Bootstrap(t *testing.T) {
	ctx := make(map[string]interface{})
	b := Bootstrapper{}

	// missing token registry
	ctx[documents.BootstrappedDocumentService] = new(testingdocuments.MockService)
	err := b.Bootstrap(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "token registry not initialisation")

	// success
	ctx[bootstrap.BootstrappedInvoiceUnpaid] = new(testingdocuments.MockRegistry)
	err = b.Bootstrap(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, ctx[extensions.BootstrappedTransferDetailService])
}
