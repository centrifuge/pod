package userapi

import (
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/extensions/transferdetails"
)

// BootstrappedUserAPIService key maps to the Service implementation in Bootstrap context.
const BootstrappedUserAPIService = "UserAPI Service"

// Bootstrapper implements bootstrap.Bootstrapper.
type Bootstrapper struct{}

// Bootstrap adds transaction.Repository into context.
func (b Bootstrapper) Bootstrap(ctx map[string]interface{}) error {
	docSrv, ok := ctx[documents.BootstrappedDocumentService].(documents.Service)
	if !ok {
		return errors.New("failed to get %s", documents.BootstrappedDocumentService)
	}

	tdSrv, ok := ctx[transferdetails.BootstrappedTransferDetailService].(transferdetails.Service)
	if !ok {
		return errors.New("failed to get %s", transferdetails.BootstrappedTransferDetailService)
	}

	ctx[BootstrappedUserAPIService] = Service{docSrv, tdSrv}
	return nil
}
