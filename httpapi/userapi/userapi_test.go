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
	assert.Len(t, r.Routes(), 16)
	assert.Equal(t, r.Routes()[0].Pattern, "/documents/{document_id}/funding_agreements")
	assert.Len(t, r.Routes()[0].Handlers, 2)
	assert.NotNil(t, r.Routes()[0].Handlers["POST"])
	assert.NotNil(t, r.Routes()[0].Handlers["GET"])
	assert.Equal(t, r.Routes()[1].Pattern, "/documents/{document_id}/funding_agreements/{agreement_id}")
	assert.Len(t, r.Routes()[1].Handlers, 1)
	assert.NotNil(t, r.Routes()[1].Handlers["GET"])
	assert.Equal(t, r.Routes()[2].Pattern, "/documents/{document_id}/transfer_details")
	assert.Len(t, r.Routes()[2].Handlers, 2)
	assert.NotNil(t, r.Routes()[2].Handlers["POST"])
	assert.NotNil(t, r.Routes()[2].Handlers["GET"])
	assert.Equal(t, r.Routes()[3].Pattern, "/documents/{document_id}/transfer_details/{transfer_id}")
	assert.Len(t, r.Routes()[3].Handlers, 2)
	assert.NotNil(t, r.Routes()[3].Handlers["PUT"])
	assert.NotNil(t, r.Routes()[3].Handlers["GET"])
	assert.Equal(t, r.Routes()[4].Pattern, "/entities")
	assert.Len(t, r.Routes()[4].Handlers, 1)
	assert.NotNil(t, r.Routes()[4].Handlers["POST"])
	assert.Equal(t, r.Routes()[5].Pattern, "/entities/{document_id}")
	assert.Len(t, r.Routes()[5].Handlers, 2)
	assert.NotNil(t, r.Routes()[5].Handlers["PUT"])
	assert.NotNil(t, r.Routes()[5].Handlers["GET"])
	assert.Equal(t, r.Routes()[6].Pattern, "/entities/{document_id}/revoke")
	assert.Len(t, r.Routes()[6].Handlers, 1)
	assert.NotNil(t, r.Routes()[6].Handlers["POST"])
	assert.Equal(t, r.Routes()[7].Pattern, "/entities/{document_id}/share")
	assert.NotNil(t, r.Routes()[7].Handlers["POST"])
	assert.Equal(t, r.Routes()[8].Pattern, "/invoices")
	assert.Len(t, r.Routes()[8].Handlers, 1)
	assert.NotNil(t, r.Routes()[8].Handlers["POST"])
	assert.Equal(t, r.Routes()[9].Pattern, "/invoices/{document_id}")
	assert.Len(t, r.Routes()[9].Handlers, 2)
	assert.NotNil(t, r.Routes()[9].Handlers["GET"])
	assert.NotNil(t, r.Routes()[9].Handlers["PUT"])
	assert.Equal(t, r.Routes()[10].Pattern, "/invoices/{document_id}/mint/unpaid")
	assert.NotNil(t, r.Routes()[10].Handlers["POST"])
	assert.Equal(t, r.Routes()[11].Pattern, "/invoices/{document_id}/versions/{version_id}")
	assert.NotNil(t, r.Routes()[11].Handlers["GET"])
	assert.Equal(t, r.Routes()[12].Pattern, "/purchase_orders")
	assert.Len(t, r.Routes()[12].Handlers, 1)
	assert.NotNil(t, r.Routes()[12].Handlers["POST"])
	assert.Equal(t, r.Routes()[13].Pattern, "/purchase_orders/{document_id}")
	assert.Len(t, r.Routes()[13].Handlers, 2)
	assert.NotNil(t, r.Routes()[13].Handlers["GET"])
	assert.NotNil(t, r.Routes()[13].Handlers["PUT"])
	assert.Equal(t, r.Routes()[14].Pattern, "/purchase_orders/{document_id}/versions/{version_id}")
	assert.NotNil(t, r.Routes()[14].Handlers["GET"])
	assert.Equal(t, r.Routes()[15].Pattern, "/relationships/{document_id}/entity")
	assert.NotNil(t, r.Routes()[15].Handlers["GET"])
}
