// +build unit

package userapi

import (
	"testing"

	"github.com/go-chi/chi"
	"github.com/stretchr/testify/assert"
)

func TestRegister(t *testing.T) {
	r := chi.NewRouter()
	Register(nil, r)
	assert.Len(t, r.Routes(), 2)
	assert.Equal(t, r.Routes()[0].Pattern, "/documents/{document_id}/transfer_details")
	assert.Len(t, r.Routes()[0].Handlers, 2)
	assert.NotNil(t, r.Routes()[0].Handlers["POST"])
	assert.NotNil(t, r.Routes()[0].Handlers["GET"])
	assert.Equal(t, r.Routes()[1].Pattern, "/documents/{document_id}/transfer_details/{transfer_id}")
	assert.Len(t, r.Routes()[1].Handlers, 2)
	assert.NotNil(t, r.Routes()[1].Handlers["PUT"])
	assert.NotNil(t, r.Routes()[1].Handlers["GET"])
}
