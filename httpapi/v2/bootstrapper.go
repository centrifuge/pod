package v2

import (
	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/oracle"
	"github.com/centrifuge/go-centrifuge/pending"
)

// BootstrappedService key maps to the Service implementation in Bootstrap context.
const BootstrappedService = "V2 Service"

// Bootstrapper implements bootstrap.Bootstrapper.
type Bootstrapper struct{}

// Bootstrap adds transaction.Repository into context.
func (b Bootstrapper) Bootstrap(ctx map[string]interface{}) error {
	pendingDocSrv, ok := ctx[pending.BootstrappedPendingDocumentService].(pending.Service)
	if !ok {
		return errors.New("failed to get %s", pending.BootstrappedPendingDocumentService)
	}

	nftSrv, ok := ctx[bootstrap.BootstrappedNFTService].(documents.TokenRegistry)
	if !ok {
		return errors.New("failed to get %s", bootstrap.BootstrappedNFTService)
	}

	oracleService, ok := ctx[oracle.BootstrappedOracleService].(oracle.Service)
	if !ok {
		return errors.New("failed to get %s", oracle.BootstrappedOracleService)
	}

	ctx[BootstrappedService] = Service{
		pendingDocSrv: pendingDocSrv,
		tokenRegistry: nftSrv,
		oracleService: oracleService,
	}
	return nil
}
