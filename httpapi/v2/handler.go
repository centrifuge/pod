package v2

import (
	"github.com/go-chi/chi"
	logging "github.com/ipfs/go-log"
)

// handler implements the API handlers.
type handler struct {
	srv Service
}

var log = logging.Logger("v2_api")

// Register registers the core apis to the router.
func Register(ctx map[string]interface{}, r chi.Router) {
	srv := ctx[BootstrappedService].(Service)
	h := handler{srv: srv}

	r.Post("/documents", h.CreateDocument)
}
