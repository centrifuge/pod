package v3

import (
	"github.com/centrifuge/go-centrifuge/http/coreapi"
	"github.com/go-chi/chi"
	logging "github.com/ipfs/go-log"
)

// handler implements the API handlers.
type handler struct {
	log *logging.ZapEventLogger
	srv Service
}

// Register registers the core apis to the router.
func Register(ctx map[string]interface{}, r chi.Router) {
	srv := ctx[BootstrappedService].(Service)
	h := handler{
		srv: srv,
		log: logging.Logger("v3_handler"),
	}

	r.Post("/nfts/classes/{"+coreapi.ClassIDParam+"}/mint", h.MintNFT)
	//r.Post("/nfts/registries/{"+coreapi.RegistryAddressParam+"}/tokens/{"+coreapi.TokenIDParam+"}/transfer", h.TransferNFT)
}
