package transferdetails

import (
	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/extensions"
	"github.com/centrifuge/go-centrifuge/httpapi"
	"github.com/centrifuge/go-centrifuge/httpapi/coreapi"
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

	tokenRegistry, ok := ctx[bootstrap.BootstrappedInvoiceUnpaid].(documents.TokenRegistry)
	if !ok {
		return errors.New("token registry not initialisation")
	}

	srv := DefaultService(func() httpapi.CoreService {
		return ctx[coreapi.BootstrappedCoreService].(httpapi.CoreService)
	}, tokenRegistry)

	ctx[extensions.BootstrappedTransferDetailService] = srv
	return nil
}
