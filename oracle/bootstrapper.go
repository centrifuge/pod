package oracle

import (
	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/ethereum"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/jobs"
	"github.com/centrifuge/go-centrifuge/queue"
)

// Bootstrapper implements bootstrap.Bootstrapper.
type Bootstrapper struct{}

// BootstrappedOracleService is the key to Oracle Service in bootstrap context.
const BootstrappedOracleService = "BootstrappedOracleService"

// Bootstrap initializes the invoice unpaid contract
func (Bootstrapper) Bootstrap(ctx map[string]interface{}) error {
	docService, ok := ctx[documents.BootstrappedDocumentService].(documents.Service)
	if !ok {
		return errors.New("document service not initialized")
	}

	idService, ok := ctx[identity.BootstrappedDIDService].(identity.Service)
	if !ok {
		return errors.New("identity service not initialised")
	}

	queueSrv, ok := ctx[bootstrap.BootstrappedQueueServer].(queue.TaskQueuer)
	if !ok {
		return errors.New("queue hasn't been initialized")
	}

	jobManager, ok := ctx[jobs.BootstrappedService].(jobs.Manager)
	if !ok {
		return errors.New("transactions repository not initialised")
	}

	client := ethereum.GetClient()
	oracleSrv := newService(
		docService,
		idService,
		client,
		queueSrv,
		jobManager)
	ctx[BootstrappedOracleService] = oracleSrv
	return nil
}
