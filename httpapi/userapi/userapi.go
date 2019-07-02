package userapi

import (
	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/go-chi/chi"
)

const (
	documentIDParam = "document_id"
	transferIDParam = "transfer_id"
)

// Register registers the core apis to the router.
func Register(ctx map[string]interface{}, r chi.Router) {
	tokenRegistry := ctx[bootstrap.BootstrappedInvoiceUnpaid].(documents.TokenRegistry)
	userAPISrv := ctx[BootstrappedUserAPIService].(Service)
	h := handler{
		tokenRegistry: tokenRegistry,
		srv:           userAPISrv,
	}
	r.Post("/documents/{"+documentIDParam+"}/transfer_details", h.CreateTransferDetail)
	r.Put("/documents/{"+documentIDParam+"}/transfer_details/{"+transferIDParam+"}", h.UpdateTransferDetail)
	r.Get("/documents/{"+documentIDParam+"}/transfer_details", h.GetTransferDetailList)
	r.Get("/documents/{"+documentIDParam+"}/transfer_details/{"+transferIDParam+"}", h.GetTransferDetail)
}
