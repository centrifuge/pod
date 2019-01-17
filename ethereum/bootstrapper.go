package ethereum

import (
	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/config/configstore"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/queue"
	"github.com/centrifuge/go-centrifuge/transactions"
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

	txService, ok := context[transactions.BootstrappedService].(transactions.Service)
	if !ok {
		return errors.New("transactions repository not initialised")
	}

	if _, ok := context[bootstrap.BootstrappedQueueServer]; !ok {
		return errors.New("queue hasn't been initialized")
	}
	queueSrv := context[bootstrap.BootstrappedQueueServer].(*queue.Server)

	client, err := NewGethClient(cfg, txService, queueSrv)
	if err != nil {
		return err
	}
	SetClient(client)
	context[BootstrappedEthereumClient] = client
	return nil
}
