package nft

import (
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/transactions"

	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/ethereum"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/queue"
)

// BootstrappedPayObService is the key to PaymentObligationService in bootstrap context.
const BootstrappedPayObService = "BootstrappedPayObService"

// Bootstrapper implements bootstrap.Bootstrapper.
type Bootstrapper struct{}

// Bootstrap initializes the payment obligation contract
func (*Bootstrapper) Bootstrap(ctx map[string]interface{}) error {
	if _, ok := ctx[bootstrap.BootstrappedConfig]; !ok {
		return errors.New("config hasn't been initialized")
	}
	cfg := ctx[bootstrap.BootstrappedConfig].(Config)

	if _, ok := ctx[ethereum.BootstrappedEthereumClient]; !ok {
		return errors.New("ethereum client hasn't been initialized")
	}

	registry, ok := ctx[documents.BootstrappedRegistry].(*documents.ServiceRegistry)
	if !ok {
		return errors.New("service registry not initialised")
	}

	idService, ok := ctx[identity.BootstrappedIDService].(identity.Service)
	if !ok {
		return errors.New("identity service not initialised")
	}

	if _, ok := ctx[bootstrap.BootstrappedQueueServer]; !ok {
		return errors.New("queue hasn't been initialized")
	}
	queueSrv := ctx[bootstrap.BootstrappedQueueServer].(*queue.Server)

	txRepository, ok := ctx[transactions.BootstrappedRepo].(transactions.Repository)
	if !ok {
		return errors.New("transactions repository not initialised")
	}

	ctx[BootstrappedPayObService] = newEthereumPaymentObligation(
		registry,
		idService,
		ethereum.GetClient(),
		cfg, queueSrv,
		bindContract,
		txRepository)

	// queue task
	task := newMintingConfirmationTask(cfg.GetEthereumContextWaitTimeout(), ethereum.DefaultWaitForTransactionMiningContext, txRepository)
	queueSrv.RegisterTaskType(task.TaskTypeName(), task)
	return nil
}
