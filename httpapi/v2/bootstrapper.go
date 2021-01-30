package v2

import (
	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/jobs/jobsv2"
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

	accountsSrv, ok := ctx[config.BootstrappedConfigStorage].(config.Service)
	if !ok {
		return errors.New("v2: failed to get accounts service")
	}

	dispatcher, ok := ctx[jobsv2.BootstrappedDispatcher].(jobsv2.Dispatcher)
	if !ok {
		return errors.New("v2: failed to get dispatcher")
	}

	ctx[BootstrappedService] = Service{
		pendingDocSrv: pendingDocSrv,
		tokenRegistry: nftSrv,
		oracleService: oracleService,
		accountSrv:    accountsSrv,
		dispatcher:    dispatcher,
	}
	return nil
}
