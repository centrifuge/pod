package userapi

import (
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/extensions/transferdetails"
)

// BootstrappedUserAPIService key maps to the service implementation in Bootstrap context.
const BootstrappedUserAPIService = "UserAPI service"

// Bootstrapper implements bootstrap.Bootstrapper.
type Bootstrapper struct{}

// Bootstrap adds transaction.Repository into context.
func (b Bootstrapper) Bootstrap(ctx map[string]interface{}) error {
	docSrv := ctx[documents.BootstrappedDocumentService].(documents.Service)
	tdSrv := ctx[transferdetails.BootstrappedTransferDetailService].(transferdetails.Service)
	ctx[BootstrappedUserAPIService] = DefaultService(docSrv, tdSrv)
	return nil
}
