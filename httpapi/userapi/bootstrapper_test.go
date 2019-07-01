// +build unit

package userapi

import (
	"testing"

	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/extensions/transferdetails"
	testingdocuments "github.com/centrifuge/go-centrifuge/testingutils/documents"
	"github.com/stretchr/testify/assert"
)

func TestBootstrapper_Bootstrap(t *testing.T) {
	ctx := make(map[string]interface{})
	b := Bootstrapper{}

	// missing doc service
	err := b.Bootstrap(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), documents.BootstrappedDocumentService)

	// missing transfer detail service
	ctx[documents.BootstrappedDocumentService] = new(testingdocuments.MockService)
	err = b.Bootstrap(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), transferdetails.BootstrappedTransferDetailService)

	// success
	ctx[transferdetails.BootstrappedTransferDetailService] = new(MockTransferService)
	assert.NoError(t, b.Bootstrap(ctx))
}
