package nft

import (
	"context"

	"github.com/centrifuge/go-centrifuge/documents/genericdoc"

	"github.com/centrifuge/go-centrifuge/config/configstore"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/identity/ethid"
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
	cfg, err := configstore.RetrieveConfig(true, ctx)
	if err != nil {
		return err
	}

	if _, ok := ctx[ethereum.BootstrappedEthereumClient]; !ok {
		return errors.New("ethereum client hasn't been initialized")
	}

	registry, ok := ctx[documents.BootstrappedRegistry].(*documents.ServiceRegistry)
	if !ok {
		return errors.New("service registry not initialised")
	}

	idService, ok := ctx[ethid.BootstrappedIDService].(identity.Service)
	if !ok {
		return errors.New("identity service not initialised")
	}

	if _, ok := ctx[bootstrap.BootstrappedQueueServer]; !ok {
		return errors.New("queue hasn't been initialized")
	}
	queueSrv := ctx[bootstrap.BootstrappedQueueServer].(*queue.Server)

	txService, ok := ctx[transactions.BootstrappedService].(transactions.Service)
	if !ok {
		return errors.New("transactions repository not initialised")
	}

	genService, ok := ctx[genericdoc.BootstrappedGenService].(genericdoc.Service)
	if !ok {
		return errors.New("generic service is not initialised")
	}

	client := ethereum.GetClient()
	ctx[BootstrappedPayObService] = newEthereumPaymentObligation(
		registry,
		idService,
		client,
		queueSrv,
		genService,
		bindContract,
		txService, func() (uint64, error) {
			h, err := client.GetEthClient().HeaderByNumber(context.Background(), nil)
			if err != nil {
				return 0, err
			}

			return h.Number.Uint64(), nil
		})

	// queue task
	task := newMintingConfirmationTask(cfg.GetEthereumContextWaitTimeout(), ethereum.DefaultWaitForTransactionMiningContext, txService)

	ethTransTask := ethereum.NewTransactionStatusTask(cfg.GetEthereumContextWaitTimeout(), txService, ethereum.DefaultWaitForTransactionMiningContext)

	queueSrv.RegisterTaskType(task.TaskTypeName(),ethTransTask)
	queueSrv.RegisterTaskType(task.TaskTypeName(), task)
	return nil
}
