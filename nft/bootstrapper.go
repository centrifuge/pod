package nft

import (
	"errors"
	"fmt"

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
	cfg := ctx[config.BootstrappedConfig].(Config)

	if _, ok := ctx[ethereum.BootstrappedEthereumClient]; !ok {
		return errors.New("ethereum client hasn't been initialized")
	}

	registry, ok := ctx[documents.BootstrappedRegistry].(*documents.ServiceRegistry)
	if !ok {
		return fmt.Errorf("service registry not initialised")
	}

	idService, ok := ctx[identity.BootstrappedIDService].(identity.Service)
	if !ok {
		return fmt.Errorf("identity service not initialised")
	}

	setPaymentObligation(NewEthereumPaymentObligation(registry, idService, ethereum.GetClient(), cfg, setupMintListener, bindContract))
	return queue.InstallQueuedTask(ctx, newMintingConfirmationTask(cfg.GetEthereumContextWaitTimeout(), ethereum.DefaultWaitForTransactionMiningContext))
}
