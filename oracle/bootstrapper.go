package oracle

import (
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/ethereum"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/jobs"
)

// Bootstrapper implements bootstrap.Bootstrapper.
type Bootstrapper struct{}

// BootstrappedOracleService is the key to Oracle Service in bootstrap context.
const BootstrappedOracleService = "BootstrappedOracleService"

// Bootstrap initializes the invoice unpaid contract
func (Bootstrapper) Bootstrap(ctx map[string]interface{}) error {
	docService := ctx[documents.BootstrappedDocumentService].(documents.Service)
	idService := ctx[identity.BootstrappedDIDService].(identity.Service)
	dispatcher := ctx[jobs.BootstrappedDispatcher].(jobs.Dispatcher)
	client := ctx[ethereum.BootstrappedEthereumClient].(ethereum.Client)
	accountSrv := ctx[config.BootstrappedConfigStorage].(config.Service)
	oracleSrv := newService(docService, idService, client, dispatcher)
	ctx[BootstrappedOracleService] = oracleSrv
	go dispatcher.RegisterRunner(oraclePushJob, &PushToOracleJob{
		accountsSrv:     accountSrv,
		identityService: idService,
		ethClient:       client,
	})
	return nil
}
