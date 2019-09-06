// +build unit

package v2

import (
	"testing"

	"github.com/go-chi/chi"
	"github.com/stretchr/testify/assert"
)

func TestRegister(t *testing.T) {
	r := chi.NewRouter()
	ctx := map[string]interface{}{BootstrappedService: Service{}}
	Register(ctx, r)
	assert.Len(t, r.Routes(), 7)
}
