package userapi

import (
	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/httpapi/coreapi"
	"github.com/go-chi/chi"
	logging "github.com/ipfs/go-log"
)

// Register registers the core apis to the router.
func Register(ctx map[string]interface{}, r chi.Router) {
	tokenRegistry := ctx[bootstrap.BootstrappedNFTService].(documents.TokenRegistry)
	userAPISrv := ctx[BootstrappedUserAPIService].(Service)
	h := handler{
		tokenRegistry: tokenRegistry,
		srv:           userAPISrv,
	}

	// entity api
	r.Post("/entities", h.CreateEntity)
	r.Put("/entities/{"+coreapi.DocumentIDParam+"}", h.UpdateEntity)
	r.Get("/entities/{"+coreapi.DocumentIDParam+"}", h.GetEntity)
	r.Post("/entities/{"+coreapi.DocumentIDParam+"}/share", h.ShareEntity)
	r.Post("/entities/{"+coreapi.DocumentIDParam+"}/revoke", h.RevokeEntity)
	r.Get("/relationships/{"+coreapi.DocumentIDParam+"}/entity", h.GetEntityThroughRelationship)
}

type handler struct {
	srv           Service
	tokenRegistry documents.TokenRegistry
}

var log = logging.Logger("user-api")
