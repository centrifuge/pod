package nft

import (
	"errors"

	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/ethereum"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/queue"

)

type Bootstrapper struct {
}

// Bootstrap initializes the payment obligation contract
func (*Bootstrapper) Bootstrap(context map[string]interface{}) error {
	if _, ok := context[bootstrap.BootstrappedConfig]; !ok {
		return errors.New("config hasn't been initialized")
	}
	cfg := context[bootstrap.BootstrappedConfig].(*config.Configuration)

	if _, ok := context[bootstrap.BootstrappedEthereumClient]; !ok {
		return errors.New("ethereum client hasn't been initialized")
	}

	setPaymentObligation(NewEthereumPaymentObligation(identity.IDService, ethereum.GetClient(), config.Config(), setupMintListener))

	contract, err := newDefaultContract()

	if err != nil {
		return err
	}

	setPaymentObligation(NewEthereumPaymentObligation(identity.IDService, ethereum.GetClient(), cfg, setupMintListener))
	return queue.InstallQueuedTask(context,
		newMintingConfirmationTask(contract, ethereum.DefaultWaitForTransactionMiningContext))
}
