package userapi

import (
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/extensions/transferdetails"
	"github.com/centrifuge/go-centrifuge/nft"
	"github.com/go-chi/chi"
)

const (
	documentIDParam = "document_id"
	transferIDParam = "transfer_id"
)

// Register registers the core apis to the router.
func Register(r chi.Router,
	nftSrv nft.Service,
	transferSrv transferdetails.Service) {
	h := handler{
		tokenRegistry: nftSrv.(documents.TokenRegistry),
		srv:           Service{transferDetailsService: transferSrv},
	}
	r.Post("/documents/{"+documentIDParam+"}/extensions/transfer_details", h.CreateTransferDetail)
	r.Post("/documents/{"+documentIDParam+"}/extensions/transfer_details/{"+transferIDParam+"}", h.UpdateTransferDetail)
	r.Get("/documents/{"+documentIDParam+"}/extensions/transfer_details", h.GetTransferDetailList)
	r.Get("/documents/{"+documentIDParam+"}/extensions/transfer_details/{"+transferIDParam+"}", h.GetTransferDetail)
	// TODO: future GET methods for specific versions
	//r.Get("/documents/{"+documentIDParam+"}/versions/{"+versionIDParam+"}/extensions/transfer_details", h.GetVersionTransferDetailList)
	//r.Get("/documents/{"+documentIDParam+"}/versions/{"+versionIDParam+"}/extensions/transfer_details/{"+transferIDParam+"}", h.GetVersionTransferDetail)
}
