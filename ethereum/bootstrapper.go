package ethereum

import (
	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/config/configstore"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/jobs"
	"github.com/centrifuge/go-centrifuge/queue"
)

// BootstrappedEthereumClient is a key to mapped client in bootstrap context.
const BootstrappedEthereumClient string = "BootstrappedEthereumClient"

// Bootstrapper implements bootstrap.Bootstrapper.
type Bootstrapper struct{}

// Bootstrap initialises ethereum client.
func (Bootstrapper) Bootstrap(context map[string]interface{}) error {
	cfg, err := configstore.RetrieveConfig(false, context)
	if err != nil {
		return err
	}

	txManager, ok := context[jobs.BootstrappedService].(jobs.Manager)
	if !ok {
		return errors.New("transactions repository not initialised")
	}

	if _, ok := context[bootstrap.BootstrappedQueueServer]; !ok {
		return errors.New("queue hasn't been initialized")
	}
	queueSrv := context[bootstrap.BootstrappedQueueServer].(*queue.Server)

	client, err := NewGethClient(cfg)
	if err != nil {
		return err
	}

	SetClient(client)
	ethTransTask := NewTransactionStatusTask(cfg.GetEthereumContextWaitTimeout(), txManager, client.TransactionByHash, client.TransactionReceipt, DefaultWaitForTransactionMiningContext)
	queueSrv.RegisterTaskType(ethTransTask.TaskTypeName(), ethTransTask)
	waitEventTask := NewWaitEventTask(txManager, DefaultWaitForTransactionMiningContext, client.GetEthClient().FilterLogs)
	queueSrv.RegisterTaskType(waitEventTask.TaskTypeName(), waitEventTask)
	context[BootstrappedEthereumClient] = client
	return nil
}
