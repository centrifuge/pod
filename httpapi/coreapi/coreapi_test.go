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
	assert.Len(t, r.Routes(), 10)
	assert.Equal(t, r.Routes()[0].Pattern, "/accounts/{account_id}/sign")
	assert.NotNil(t, r.Routes()[0].Handlers["POST"])
	assert.Equal(t, r.Routes()[1].Pattern, "/documents")
	assert.Len(t, r.Routes()[1].Handlers, 2)
	assert.NotNil(t, r.Routes()[1].Handlers["POST"])
	assert.NotNil(t, r.Routes()[1].Handlers["PUT"])
	assert.Equal(t, r.Routes()[2].Pattern, "/documents/{document_id}")
	assert.NotNil(t, r.Routes()[2].Handlers["GET"])
	assert.Equal(t, r.Routes()[3].Pattern, "/documents/{document_id}/proofs")
	assert.NotNil(t, r.Routes()[3].Handlers["POST"])
	assert.Equal(t, r.Routes()[4].Pattern, "/documents/{document_id}/versions/{version_id}")
	assert.NotNil(t, r.Routes()[4].Handlers["GET"])
	assert.Equal(t, r.Routes()[5].Pattern, "/documents/{document_id}/versions/{version_id}/proofs")
	assert.NotNil(t, r.Routes()[5].Handlers["POST"])
	assert.Equal(t, r.Routes()[6].Pattern, "/jobs/{job_id}")
	assert.NotNil(t, r.Routes()[6].Handlers["GET"])
	assert.Equal(t, r.Routes()[7].Pattern, "/nfts/mint")
	assert.NotNil(t, r.Routes()[7].Handlers["POST"])
	assert.Equal(t, r.Routes()[8].Pattern, "/nfts/{token_id}/registry/{registry_address}/owner")
	assert.NotNil(t, r.Routes()[8].Handlers["GET"])
	assert.Equal(t, r.Routes()[9].Pattern, "/nfts/{token_id}/transfer")
	assert.NotNil(t, r.Routes()[9].Handlers["POST"])
}
