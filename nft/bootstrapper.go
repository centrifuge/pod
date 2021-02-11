package nft

import (
	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/centchain"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/ethereum"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/jobs"
)

// Bootstrapper implements bootstrap.Bootstrapper.
type Bootstrapper struct{}

// Bootstrap initializes the invoice unpaid contract
func (*Bootstrapper) Bootstrap(ctx map[string]interface{}) error {
	centAPI, ok := ctx[centchain.BootstrappedCentChainClient].(centchain.API)
	if !ok {
		return errors.New("centchain client hasn't been initialized")
	}

	docSrv, ok := ctx[documents.BootstrappedDocumentService].(documents.Service)
	if !ok {
		return errors.New("document service not initialised")
	}

	idService, ok := ctx[identity.BootstrappedDIDService].(identity.Service)
	if !ok {
		return errors.New("identity service not initialised")
	}

	accountsSrv := ctx[config.BootstrappedConfigStorage].(config.Service)
	dispatcher := ctx[jobs.BootstrappedDispatcher].(jobs.Dispatcher)
	ethClient := ctx[ethereum.BootstrappedEthereumClient].(ethereum.Client)
	api := api{api: centAPI}
	go dispatcher.RegisterRunner(nftJob, &MintNFTJob{
		accountsSrv: accountsSrv,
		docSrv:      docSrv,
		dispatcher:  dispatcher,
		ethClient:   ethClient,
		api:         api,
		identitySrv: idService,
	})

	go dispatcher.RegisterRunner(transferNFTJob, &TransferNFTJob{
		identitySrv: idService,
		accountSrv:  accountsSrv,
		ethClient:   ethClient,
	})

	nftSrv := newService(
		ethClient,
		docSrv,
		ethereum.BindContract,
		dispatcher)
	ctx[bootstrap.BootstrappedNFTService] = nftSrv
	return nil
}
