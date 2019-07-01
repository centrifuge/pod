package transferdetails

import (
	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/httpapi/coreapi"
)

const (
	// BootstrappedTransferDetailService is the key to bootstrapped document service
	BootstrappedTransferDetailService = "BootstrappedTransferDetailsService"
)

// Bootstrapper implements Bootstrapper Interface
type Bootstrapper struct{}

// Bootstrap adds the funding API handler to the context.
func (Bootstrapper) Bootstrap(ctx map[string]interface{}) (err error) {
	defer func() {
		if err != nil {
			err = errors.New("transferdetail bootstrapper: %v", err)
		}
	}()

	coreAPISrv, ok := ctx[coreapi.BootstrappedCoreAPIService].(coreapi.Service)
	if !ok {
		return errors.New("core-api service not initialised")
	}

	tokenRegistry, ok := ctx[bootstrap.BootstrappedInvoiceUnpaid].(documents.TokenRegistry)
	if !ok {
		return errors.New("token registry not initialisation")
	}

	srv := DefaultService(coreAPISrv, tokenRegistry)
	ctx[BootstrappedTransferDetailService] = srv
	return nil
}
