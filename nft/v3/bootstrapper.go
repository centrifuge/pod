package v3

import (
	"github.com/centrifuge/pod/config"
	"github.com/centrifuge/pod/documents"
	"github.com/centrifuge/pod/errors"
	"github.com/centrifuge/pod/ipfs"
	"github.com/centrifuge/pod/jobs"
	"github.com/centrifuge/pod/pallets"
	"github.com/centrifuge/pod/pallets/uniques"
	"github.com/centrifuge/pod/pallets/utility"
	"github.com/centrifuge/pod/pending"
	"github.com/centrifuge/pod/storage"
)

const (
	BootstrappedNFTV3Service = "BootstrappedNFTV3Service"
)

type Bootstrapper struct{}

func (*Bootstrapper) Bootstrap(ctx map[string]interface{}) error {
	accountsSrv, ok := ctx[config.BootstrappedConfigStorage].(config.Service)

	if !ok {
		return errors.New("config storage not initialised")
	}

	docSrv, ok := ctx[documents.BootstrappedDocumentService].(documents.Service)

	if !ok {
		return errors.New("documents service not initialised")
	}

	pendingDocsSrv, ok := ctx[pending.BootstrappedPendingDocumentService].(pending.Service)

	if !ok {
		return errors.New("pending documents service not initialised")
	}

	dispatcher, ok := ctx[jobs.BootstrappedJobDispatcher].(jobs.Dispatcher)

	if !ok {
		return errors.New("jobs dispatcher not initialised")
	}

	ipfsPinningSrv, ok := ctx[ipfs.BootstrappedIPFSPinningService].(ipfs.PinningServiceClient)

	if !ok {
		return errors.New("ipfs pinning service not initialised")
	}

	uniquesAPI, ok := ctx[pallets.BootstrappedUniquesAPI].(uniques.API)

	if !ok {
		return errors.New("proxy API not initialised")
	}

	utilityAPI, ok := ctx[pallets.BootstrappedUtilityAPI].(utility.API)

	if !ok {
		return errors.New("utility API not initialised")
	}

	ldb, ok := ctx[storage.BootstrappedDB].(storage.Repository)

	if !ok {
		return errors.New("DB not found in the bootstrapper")
	}

	repo := pending.NewRepository(ldb)

	go dispatcher.RegisterRunner(mintNFTForPendingDocV3Job, &MintNFTForPendingDocJobRunner{
		accountsSrv:    accountsSrv,
		pendingDocsSrv: pendingDocsSrv,
		pendingRepo:    repo,
		docSrv:         docSrv,
		dispatcher:     dispatcher,
		utilityAPI:     utilityAPI,
		ipfsPinningSrv: ipfsPinningSrv,
	})

	go dispatcher.RegisterRunner(mintNFTForCommittedDocV3Job, &MintNFTForCommittedDocJobRunner{
		accountsSrv:    accountsSrv,
		docSrv:         docSrv,
		dispatcher:     dispatcher,
		utilityAPI:     utilityAPI,
		ipfsPinningSrv: ipfsPinningSrv,
	})

	go dispatcher.RegisterRunner(createNFTCollectionV3Job, &CreateCollectionJobRunner{
		accountsSrv: accountsSrv,
		docSrv:      docSrv,
		dispatcher:  dispatcher,
		api:         uniquesAPI,
	})

	nftService := NewService(
		pendingDocsSrv,
		docSrv,
		dispatcher,
		uniquesAPI,
	)

	ctx[BootstrappedNFTV3Service] = nftService

	return nil
}
