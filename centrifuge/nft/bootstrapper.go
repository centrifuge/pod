package nft

import (
	"github.com/centrifuge/go-centrifuge/centrifuge/bootstrap"
		"errors"
	"github.com/centrifuge/go-centrifuge/centrifuge/ethereum"
	"github.com/centrifuge/go-centrifuge/centrifuge/config"
	"github.com/centrifuge/go-centrifuge/centrifuge/identity"
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
	// TODO check if identity has been initialized
	contract, err := getPaymentObligationContract()
	if err != nil {
		return err
	}
	setPaymentObligationService(NewPaymentObligationService(contract, identity.IDService, config.Config))
	return nil
}

func (b *Bootstrapper) TestBootstrap(context map[string]interface{}) error {
	return b.Bootstrap(context)
}

func getPaymentObligationContract() (*EthereumPaymentObligationContract, error) {
	client := ethereum.GetConnection()
	return NewEthereumPaymentObligationContract(config.Config.GetContractAddress("paymentObligation"), client.GetClient())
}
