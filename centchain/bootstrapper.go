package centchain

import (
	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/config/configstore"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/jobs"
	"github.com/centrifuge/go-centrifuge/queue"
	gsrpc "github.com/centrifuge/go-substrate-rpc-client"
)

// BootstrappedCentChainClient is a key to mapped client in bootstrap context.
const BootstrappedCentChainClient string = "BootstrappedCentChainClient"

// Bootstrapper implements bootstrap.Bootstrapper.
type Bootstrapper struct{}

// Bootstrap initialises centchain client.
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

	sapi, err := gsrpc.NewSubstrateAPI(cfg.GetCentChainNodeURL())
	if err != nil {
		return err
	}

	client := NewAPI(sapi, cfg, queueSrv)
	extStatusTask := NewExtrinsicStatusTask(cfg.GetEthereumContextWaitTimeout(), cfg.GetCentChainMaxRetries(), txManager, sapi.RPC.Chain.GetBlockHash, sapi.RPC.Chain.GetBlock, client.GetMetadataLatest, sapi.RPC.State.GetStorage)
	queueSrv.RegisterTaskType(extStatusTask.TaskTypeName(), extStatusTask)
	context[BootstrappedCentChainClient] = client

	return nil
}
