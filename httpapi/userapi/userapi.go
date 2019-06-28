package userapi

import (
	"context"

	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/extensions"
	"github.com/centrifuge/go-centrifuge/httpapi"
	"github.com/centrifuge/go-centrifuge/httpapi/coreapi"
	"github.com/centrifuge/go-centrifuge/nft"
	"github.com/go-chi/chi"
)

const (
	documentIDParam = "document_id"
	transferIDParam = "transfer_id"
)

// Register registers the core apis to the router.
func Register(ctx context.Context, r chi.Router) error {
	// node object registry
	nodeObjReg, ok := ctx.Value(bootstrap.NodeObjRegistry).(map[string]interface{})
	if !ok {
		return errors.New("failed to get %s", bootstrap.NodeObjRegistry)
	}

	nftSrv, ok := nodeObjReg[bootstrap.BootstrappedInvoiceUnpaid].(nft.InvoiceUnpaid)
	if !ok {
		return errors.New("failed to get %s", bootstrap.BootstrappedInvoiceUnpaid)
	}

	docSrv, ok := nodeObjReg[coreapi.BootstrappedCoreService].(httpapi.CoreService)
	if !ok {
		return errors.New("failed to get %s", documents.BootstrappedDocumentService)
	}

	transferSrv, ok := nodeObjReg[extensions.BootstrappedTransferDetailService].(extensions.TransferDetailService)
	if !ok {
		return errors.New("failed to get %s", extensions.BootstrappedTransferDetailService)
	}

	h := handler{
		tokenRegistry: nftSrv.(documents.TokenRegistry),
		srv:           Service{coreService: docSrv, transferDetailsService: transferSrv},
	}
	r.Post("/documents/{"+documentIDParam+"}/transfer_details", h.CreateTransferDetail)
	r.Put("/documents/{"+documentIDParam+"}/transfer_details/{"+transferIDParam+"}", h.UpdateTransferDetail)
	r.Get("/documents/{"+documentIDParam+"}/transfer_details", h.GetTransferDetailList)
	r.Get("/documents/{"+documentIDParam+"}/transfer_details/{"+transferIDParam+"}", h.GetTransferDetail)

	return nil
}
