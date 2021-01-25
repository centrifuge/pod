// +build unit

package userapi

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
		BootstrappedUserAPIService:       Service{},
		bootstrap.BootstrappedNFTService: new(testingnfts.MockNFTService),
	}
	Register(ctx, r)
	assert.Len(t, r.Routes(), 5)
}
