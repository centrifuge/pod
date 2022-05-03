package v3

import (
	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/centchain"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/jobs"
)

type Bootstrapper struct{}

func (*Bootstrapper) Bootstrap(ctx map[string]interface{}) error {
	accountsSrv := ctx[config.BootstrappedConfigStorage].(config.Service)
	centAPI := ctx[centchain.BootstrappedCentChainClient].(centchain.API)
	docSrv := ctx[documents.BootstrappedDocumentService].(documents.Service)
	dispatcher := ctx[jobs.BootstrappedDispatcher].(jobs.Dispatcher)
	uniquesAPI := newUniquesAPI(centAPI)

	go dispatcher.RegisterRunner(mintNFTV3Job, &MintNFTJob{
		accountsSrv: accountsSrv,
		docSrv:      docSrv,
		dispatcher:  dispatcher,
		api:         uniquesAPI,
	})

	nftService := newService(
		docSrv,
		dispatcher,
		uniquesAPI,
	)

	ctx[bootstrap.BootstrappedNFTV3Service] = nftService

	return nil
}
