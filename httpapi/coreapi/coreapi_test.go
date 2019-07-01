// +build unit

package coreapi

import (
	"testing"

	"github.com/centrifuge/go-centrifuge/bootstrap"
	testingnfts "github.com/centrifuge/go-centrifuge/testingutils/nfts"
	"github.com/go-chi/chi"
	"github.com/stretchr/testify/assert"
)

func TestRegister(t *testing.T) {
	r := chi.NewRouter()
	ctx := map[string]interface{}{
		BootstrappedCoreAPIService:          Service{},
		bootstrap.BootstrappedInvoiceUnpaid: new(testingnfts.MockNFTService),
	}
	Register(ctx, r)
	assert.Len(t, r.Routes(), 10)
	assert.Equal(t, r.Routes()[0].Pattern, "/accounts/{account_id}/sign")
	assert.NotNil(t, r.Routes()[0].Handlers["POST"])
	assert.Equal(t, r.Routes()[1].Pattern, "/documents")
	assert.NotNil(t, r.Routes()[1].Handlers["POST"])
	assert.Equal(t, r.Routes()[2].Pattern, "/documents/{document_id}")
	assert.Len(t, r.Routes()[2].Handlers, 2)
	assert.NotNil(t, r.Routes()[2].Handlers["GET"])
	assert.NotNil(t, r.Routes()[2].Handlers["PUT"])
	assert.Equal(t, r.Routes()[3].Pattern, "/documents/{document_id}/proofs")
	assert.NotNil(t, r.Routes()[3].Handlers["POST"])
	assert.Equal(t, r.Routes()[4].Pattern, "/documents/{document_id}/versions/{version_id}")
	assert.NotNil(t, r.Routes()[4].Handlers["GET"])
	assert.Equal(t, r.Routes()[5].Pattern, "/documents/{document_id}/versions/{version_id}/proofs")
	assert.NotNil(t, r.Routes()[5].Handlers["POST"])
	assert.Equal(t, r.Routes()[6].Pattern, "/jobs/{job_id}")
	assert.NotNil(t, r.Routes()[6].Handlers["GET"])
	assert.Equal(t, r.Routes()[7].Pattern, "/nfts/registries/{registry_address}/mint")
	assert.NotNil(t, r.Routes()[7].Handlers["POST"])
	assert.Equal(t, r.Routes()[8].Pattern, "/nfts/registries/{registry_address}/tokens/{token_id}/owner")
	assert.NotNil(t, r.Routes()[8].Handlers["GET"])
	assert.Equal(t, r.Routes()[9].Pattern, "/nfts/registries/{registry_address}/tokens/{token_id}/transfer")
	assert.NotNil(t, r.Routes()[9].Handlers["POST"])
}
