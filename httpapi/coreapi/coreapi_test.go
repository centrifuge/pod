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
	Register(r, new(testingnfts.MockNFTService), nil, nil)
	assert.Len(t, r.Routes(), 7)
	assert.Equal(t, r.Routes()[0].Pattern, "/documents")
	assert.Len(t, r.Routes()[0].Handlers, 2)
	assert.NotNil(t, r.Routes()[0].Handlers["POST"])
	assert.NotNil(t, r.Routes()[0].Handlers["PUT"])
	assert.Equal(t, r.Routes()[1].Pattern, "/documents/{document_id}")
	assert.NotNil(t, r.Routes()[1].Handlers["GET"])
	assert.Equal(t, r.Routes()[2].Pattern, "/documents/{document_id}/proofs")
	assert.NotNil(t, r.Routes()[2].Handlers["POST"])
	assert.Equal(t, r.Routes()[3].Pattern, "/documents/{document_id}/versions/{version_id}")
	assert.NotNil(t, r.Routes()[3].Handlers["GET"])
	assert.Equal(t, r.Routes()[4].Pattern, "/documents/{document_id}/versions/{version_id}/proofs")
	assert.NotNil(t, r.Routes()[4].Handlers["POST"])
	assert.Equal(t, r.Routes()[5].Pattern, "/jobs/{job_id}")
	assert.NotNil(t, r.Routes()[5].Handlers["GET"])
	assert.Equal(t, r.Routes()[6].Pattern, "/nfts/mint")
	assert.NotNil(t, r.Routes()[6].Handlers["POST"])
}
