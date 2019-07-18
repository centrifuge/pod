package funding

import (
	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/errors"
)

const (
	// BootstrappedFundingService is the key for Funding service in Context.
	BootstrappedFundingService = "Funding Service"
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

	docSrv, ok := ctx[documents.BootstrappedDocumentService].(documents.Service)
	if !ok {
		return errors.New("document service not initialised")
	}

	tokenRegistry, ok := ctx[bootstrap.BootstrappedInvoiceUnpaid].(documents.TokenRegistry)
	if !ok {
		return errors.New("token registry not initialisation")
	}

	srv := DefaultService(docSrv, tokenRegistry)
	ctx[BootstrappedFundingService] = srv
	return nil
}
