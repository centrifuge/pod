// +build unit

package coreapi

import (
	"testing"

	testingnfts "github.com/centrifuge/go-centrifuge/testingutils/nfts"
	"github.com/go-chi/chi"
	"github.com/stretchr/testify/assert"
)

func TestRegister(t *testing.T) {
	r := chi.NewRouter()
	Register(r, new(testingnfts.MockNFTService), nil, nil, nil)
	assert.Len(t, r.Routes(), 11)
	assert.Equal(t, r.Routes()[0].Pattern, "/accounts/{account_id}")
	assert.NotNil(t, r.Routes()[0].Handlers["GET"])
	assert.Equal(t, r.Routes()[1].Pattern, "/accounts/{account_id}/sign")
	assert.NotNil(t, r.Routes()[1].Handlers["POST"])
	assert.Equal(t, r.Routes()[2].Pattern, "/documents")
	assert.NotNil(t, r.Routes()[2].Handlers["POST"])
	assert.Equal(t, r.Routes()[3].Pattern, "/documents/{document_id}")
	assert.Len(t, r.Routes()[3].Handlers, 2)
	assert.NotNil(t, r.Routes()[3].Handlers["GET"])
	assert.NotNil(t, r.Routes()[3].Handlers["PUT"])
	assert.Equal(t, r.Routes()[4].Pattern, "/documents/{document_id}/proofs")
	assert.NotNil(t, r.Routes()[4].Handlers["POST"])
	assert.Equal(t, r.Routes()[5].Pattern, "/documents/{document_id}/versions/{version_id}")
	assert.NotNil(t, r.Routes()[5].Handlers["GET"])
	assert.Equal(t, r.Routes()[6].Pattern, "/documents/{document_id}/versions/{version_id}/proofs")
	assert.NotNil(t, r.Routes()[6].Handlers["POST"])
	assert.Equal(t, r.Routes()[7].Pattern, "/jobs/{job_id}")
	assert.NotNil(t, r.Routes()[7].Handlers["GET"])
	assert.Equal(t, r.Routes()[8].Pattern, "/nfts/registries/{registry_address}/mint")
	assert.NotNil(t, r.Routes()[8].Handlers["POST"])
	assert.Equal(t, r.Routes()[9].Pattern, "/nfts/registries/{registry_address}/tokens/{token_id}/owner")
	assert.NotNil(t, r.Routes()[9].Handlers["GET"])
	assert.Equal(t, r.Routes()[10].Pattern, "/nfts/registries/{registry_address}/tokens/{token_id}/transfer")
	assert.NotNil(t, r.Routes()[10].Handlers["POST"])
}
