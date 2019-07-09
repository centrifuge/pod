package userapi

import (
	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/httpapi/coreapi"
	"github.com/go-chi/chi"
)

const (
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

	// transfer details api
	r.Post("/documents/{"+coreapi.DocumentIDParam+"}/transfer_details", h.CreateTransferDetail)
	r.Put("/documents/{"+coreapi.DocumentIDParam+"}/transfer_details/{"+transferIDParam+"}", h.UpdateTransferDetail)
	r.Get("/documents/{"+coreapi.DocumentIDParam+"}/transfer_details", h.GetTransferDetailList)
	r.Get("/documents/{"+coreapi.DocumentIDParam+"}/transfer_details/{"+transferIDParam+"}", h.GetTransferDetail)

	// invoice apis
	r.Post("/invoices", h.CreateInvoice)
	r.Get("/invoices/{"+coreapi.DocumentIDParam+"}", h.GetInvoice)
	r.Put("/invoices/{"+coreapi.DocumentIDParam+"}", h.UpdateInvoice)

	// purchase order api
	r.Post("/purchase_orders", h.CreatePurchaseOrder)
	r.Get("/purchase_orders/{"+coreapi.DocumentIDParam+"}", h.GetPurchaseOrder)
	r.Put("/purchase_orders/{"+coreapi.DocumentIDParam+"}", h.UpdatePurchaseOrder)
	r.Get("/purchase_orders/{"+coreapi.DocumentIDParam+"}/versions/{"+coreapi.VersionIDParam+"}", h.GetPurchaseOrderVersion)

	// entity api
	r.Post("/entities", h.CreateEntity)
	r.Put("/entities/{"+coreapi.DocumentIDParam+"}", h.UpdateEntity)
	r.Get("/entities/{"+coreapi.DocumentIDParam+"}", h.GetEntity)
}
