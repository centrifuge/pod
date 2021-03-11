package nft

import (
	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/centchain"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/ethereum"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/jobs"
)

// Bootstrapper implements bootstrap.Bootstrapper.
type Bootstrapper struct{}

// Bootstrap initializes the invoice unpaid contract
func (*Bootstrapper) Bootstrap(ctx map[string]interface{}) error {
	centAPI := ctx[centchain.BootstrappedCentChainClient].(centchain.API)
	docSrv := ctx[documents.BootstrappedDocumentService].(documents.Service)
	idService := ctx[identity.BootstrappedDIDService].(identity.Service)
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
