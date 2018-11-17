package nft

import (
	"errors"
	"fmt"

	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/ethereum"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/queue"
)

type Bootstrapper struct {
}

// Bootstrap initializes the payment obligation contract
func (*Bootstrapper) Bootstrap(ctx map[string]interface{}) error {
	if _, ok := ctx[config.BootstrappedConfig]; !ok {
		return errors.New("config hasn't been initialized")
	}
	cfg := ctx[config.BootstrappedConfig].(*config.Configuration)

	if _, ok := ctx[ethereum.BootstrappedEthereumClient]; !ok {
		return errors.New("ethereum client hasn't been initialized")
	}

	registry, ok := ctx[documents.BootstrappedRegistry].(*documents.ServiceRegistry)
	if !ok {
		return fmt.Errorf("service registry not initialised")
	}

	setPaymentObligation(NewEthereumPaymentObligation(registry, identity.IDService, ethereum.GetClient(), cfg, setupMintListener, bindContract))
	return queue.InstallQueuedTask(ctx, newMintingConfirmationTask(ethereum.DefaultWaitForTransactionMiningContext))
}
