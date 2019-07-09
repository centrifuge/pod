// +build unit

package userapi

import (
	"testing"

	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/testingutils/nfts"
	"github.com/go-chi/chi"
	"github.com/stretchr/testify/assert"
)

func TestRegister(t *testing.T) {
	r := chi.NewRouter()
	ctx := map[string]interface{}{
		BootstrappedUserAPIService:          Service{},
		bootstrap.BootstrappedInvoiceUnpaid: new(testingnfts.MockNFTService),
	}
	Register(ctx, r)
	assert.Len(t, r.Routes(), 8)
	assert.Equal(t, r.Routes()[0].Pattern, "/documents/{document_id}/transfer_details")
	assert.Len(t, r.Routes()[0].Handlers, 2)
	assert.NotNil(t, r.Routes()[0].Handlers["POST"])
	assert.NotNil(t, r.Routes()[0].Handlers["GET"])
	assert.Equal(t, r.Routes()[1].Pattern, "/documents/{document_id}/transfer_details/{transfer_id}")
	assert.Len(t, r.Routes()[1].Handlers, 2)
	assert.NotNil(t, r.Routes()[1].Handlers["PUT"])
	assert.NotNil(t, r.Routes()[1].Handlers["GET"])
	assert.Equal(t, r.Routes()[2].Pattern, "/entities")
	assert.Len(t, r.Routes()[2].Handlers, 1)
	assert.NotNil(t, r.Routes()[2].Handlers["POST"])

	assert.Equal(t, r.Routes()[3].Pattern, "/invoices")
	assert.Len(t, r.Routes()[3].Handlers, 1)
	assert.NotNil(t, r.Routes()[3].Handlers["POST"])
	assert.Equal(t, r.Routes()[4].Pattern, "/invoices/{document_id}")
	assert.Len(t, r.Routes()[4].Handlers, 2)
	assert.NotNil(t, r.Routes()[4].Handlers["GET"])
	assert.NotNil(t, r.Routes()[4].Handlers["PUT"])

	assert.Equal(t, r.Routes()[5].Pattern, "/purchase_orders")
	assert.Len(t, r.Routes()[5].Handlers, 1)
	assert.NotNil(t, r.Routes()[5].Handlers["POST"])
	assert.Equal(t, r.Routes()[6].Pattern, "/purchase_orders/{document_id}")
	assert.Len(t, r.Routes()[6].Handlers, 2)
	assert.NotNil(t, r.Routes()[6].Handlers["GET"])
	assert.NotNil(t, r.Routes()[6].Handlers["PUT"])
	assert.Equal(t, r.Routes()[7].Pattern, "/purchase_orders/{document_id}/versions/{version_id}")
	assert.NotNil(t, r.Routes()[7].Handlers["GET"])
}
