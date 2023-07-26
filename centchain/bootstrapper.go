package centchain

import (
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/jobs"
	gsrpc "github.com/centrifuge/go-substrate-rpc-client/v4"
	"github.com/centrifuge/go-substrate-rpc-client/v4/registry/retriever"
	"github.com/centrifuge/go-substrate-rpc-client/v4/registry/state"
)

// BootstrappedCentChainClient is a key to mapped client in bootstrap context.
const BootstrappedCentChainClient string = "BootstrappedCentChainClient"

// Bootstrapper implements bootstrap.Bootstrapper.
type Bootstrapper struct{}

// Bootstrap initialises centchain client.
func (Bootstrapper) Bootstrap(context map[string]interface{}) error {
	cfg, err := config.RetrieveConfig(false, context)
	if err != nil {
		return err
	}

	dispatcher := context[jobs.BootstrappedDispatcher].(jobs.Dispatcher)
	sapi, err := gsrpc.NewSubstrateAPI(cfg.GetCentChainNodeURL())
	if err != nil {
		return err
	}
	eventRetriever, err := retriever.NewDefaultEventRetriever(state.NewEventProvider(sapi.RPC.State), sapi.RPC.State)
	if err != nil {
		return err
	}
	centSAPI := &defaultSubstrateAPI{sapi}
	client := NewAPI(centSAPI, cfg, dispatcher, eventRetriever)
	context[BootstrappedCentChainClient] = client
	return nil
}
