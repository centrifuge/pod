package userapi

import (
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/extensions/transferdetails"
	"github.com/centrifuge/go-centrifuge/nft"
	"github.com/go-chi/chi"
)

// Register registers the core apis to the router.
func Register(r chi.Router,
	nftSrv nft.Service,
	transferSrv transferdetails.Service) {
	h := handler{
		tokenRegistry: nftSrv.(documents.TokenRegistry),
		srv:           Service{transferDetailsService: transferSrv},
	}
	r.Post("/extensions/", h.CreateTransferDetail)
}