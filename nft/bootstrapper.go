package nft

import (
	"errors"

	"github.com/centrifuge/go-centrifuge/queue"
	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/ethereum"
	"github.com/centrifuge/go-centrifuge/identity"
)

type Bootstrapper struct {
}

// Bootstrap initializes the payment obligation contract
func (*Bootstrapper) Bootstrap(context map[string]interface{}) error {
	if _, ok := context[bootstrap.BootstrappedConfig]; !ok {
		return errors.New("config hasn't been initialized")
	}
	if _, ok := context[bootstrap.BootstrappedEthereumClient]; !ok {
		return errors.New("ethereum client hasn't been initialized")
	}

	contract, err := getPaymentObligationContract()
	if err != nil {
		return err
	}

	setPaymentObligation(NewEthereumPaymentObligation(contract, identity.IDService, ethereum.GetClient(), config.Config(),setupMintListener))
	return queue.InstallQueuedTask(context,
		NewMintingConfirmationTask(contract, ethereum.DefaultWaitForTransactionMiningContext))
}

func getPaymentObligationContract() (*EthereumPaymentObligationContract, error) {
	client := ethereum.GetClient()
	return NewEthereumPaymentObligationContract(config.Config().GetContractAddress("paymentObligation"), client.GetEthClient())
}
