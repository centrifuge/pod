package coreapi

import (
	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/jobs"
	"github.com/centrifuge/go-centrifuge/nft"
)

// BootstrappedCoreAPIService key maps to the Service implementation in Bootstrap context.
const BootstrappedCoreAPIService = "CoreAPI Service"

// Bootstrapper implements bootstrap.Bootstrapper.
type Bootstrapper struct{}

// Bootstrap adds transaction.Repository into context.
func (b Bootstrapper) Bootstrap(ctx map[string]interface{}) error {
	docSrv, ok := ctx[documents.BootstrappedDocumentService].(documents.Service)
	if !ok {
		return errors.New("failed to get %s", documents.BootstrappedDocumentService)
	}

	jobsMan, ok := ctx[jobs.BootstrappedService].(jobs.Manager)
	if !ok {
		return errors.New("failed to get %s", jobs.BootstrappedService)
	}

	nftSrv, ok := ctx[bootstrap.BootstrappedNFTService].(nft.Service)
	if !ok {
		return errors.New("failed to get %s", bootstrap.BootstrappedNFTService)
	}

	accountSrv, ok := ctx[config.BootstrappedConfigStorage].(config.Service)
	if !ok {
		return errors.New("failed to get %s", config.BootstrappedConfigStorage)
	}

	ctx[BootstrappedCoreAPIService] = Service{
		docSrv:      docSrv,
		jobsSrv:     jobsMan,
		nftSrv:      nftSrv,
		accountsSrv: accountSrv,
	}
	return nil
}
