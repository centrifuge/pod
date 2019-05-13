package funding

import (
	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/identity"
)

const (
	// BootstrappedFundingAPIHandler is the key for the api handler in Context
	BootstrappedFundingAPIHandler = "Funding API Handler"
)

// Bootstrapper implements Bootstrapper Interface
type Bootstrapper struct{}

// Bootstrap adds the funding API handler to the context.
func (Bootstrapper) Bootstrap(ctx map[string]interface{}) (err error) {
	defer func() {
		if err != nil {
			err = errors.New("funding bootstrapper: %v", err)
		}
	}()

	cfgSrv, ok := ctx[config.BootstrappedConfigStorage].(config.Service)
	if !ok {
		return errors.New("config service not initialised")
	}

	docSrv, ok := ctx[documents.BootstrappedDocumentService].(documents.Service)
	if !ok {
		return errors.New("document service not initialised")
	}
	idSrv, ok := ctx[identity.BootstrappedDIDService].(identity.Service)
	if !ok {
		return errors.New("identity service not initialized")
	}

	tokenRegistry, ok := ctx[bootstrap.BootstrappedInvoiceUnpaid].(documents.TokenRegistry)
	if !ok {
		return errors.New("token registry not initialisation")
	}



	srv := DefaultService(docSrv, tokenRegistry, idSrv)
	handler := GRPCHandler(cfgSrv, srv)
	ctx[BootstrappedFundingAPIHandler] = handler
	return nil
}
