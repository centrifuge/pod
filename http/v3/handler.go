package v3

import (
	"github.com/centrifuge/pod/http/coreapi"
	"github.com/go-chi/chi"
	logging "github.com/ipfs/go-log"
)

// handler implements the API handlers.
type handler struct {
	log *logging.ZapEventLogger
	srv *Service
}

// Register registers the core apis to the router.
func Register(ctx map[string]interface{}, r chi.Router) {
	srv := ctx[BootstrappedService].(*Service)
	h := handler{
		srv: srv,
		log: logging.Logger("v3_api"),
	}

	r.Post("/nfts/collections", h.CreateNFTCollection)
	r.Post("/nfts/collections/{"+coreapi.CollectionIDParam+"}/mint", h.MintNFT)
	r.Post("/nfts/collections/{"+coreapi.CollectionIDParam+"}/commit_and_mint", h.CommitAndMintNFT)
	r.Get("/nfts/collections/{"+coreapi.CollectionIDParam+"}/items/{"+coreapi.ItemIDParam+"}/owner", h.GetNFTOwner)
	r.Get("/nfts/collections/{"+coreapi.CollectionIDParam+"}/items/{"+coreapi.ItemIDParam+"}/metadata", h.MetadataOfNFT)
	r.Get("/nfts/collections/{"+coreapi.CollectionIDParam+"}/items/{"+coreapi.ItemIDParam+"}/attribute/{"+coreapi.AttributeNameParam+"}", h.AttributeOfNFT)
	r.Get("/investor/assets", h.GetAsset)
}
