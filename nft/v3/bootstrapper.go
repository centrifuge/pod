package v3

import (
	"github.com/centrifuge/go-centrifuge/centchain"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/errors"
	v2 "github.com/centrifuge/go-centrifuge/identity/v2"
	v2proxy "github.com/centrifuge/go-centrifuge/identity/v2/proxy"
	"github.com/centrifuge/go-centrifuge/ipfs_pinning"
	"github.com/centrifuge/go-centrifuge/jobs"
	"github.com/centrifuge/go-centrifuge/nft/v3/uniques"
	"github.com/centrifuge/go-centrifuge/pending"
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

	centAPI, ok := ctx[centchain.BootstrappedCentChainClient].(centchain.API)

	if !ok {
		return errors.New("centchain API not initialised")
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

	ipfsPinningSrv, ok := ctx[ipfs_pinning.BootstrappedIPFSPinningService].(ipfs_pinning.PinningServiceClient)

	if !ok {
		return errors.New("ipfs pinning service not initialised")
	}

	proxyAPI, ok := ctx[v2.BootstrappedProxyAPI].(v2proxy.API)

	if !ok {
		return errors.New("proxy API not initialised")
	}

	uniquesAPI := uniques.NewAPI(centAPI, proxyAPI)

	go dispatcher.RegisterRunner(commitAndMintNFTV3Job, &CommitAndMintNFTJobRunner{
		accountsSrv:    accountsSrv,
		pendingDocsSrv: pendingDocsSrv,
		docSrv:         docSrv,
		dispatcher:     dispatcher,
		api:            uniquesAPI,
		ipfsPinningSrv: ipfsPinningSrv,
	})

	go dispatcher.RegisterRunner(mintNFTV3Job, &MintNFTJobRunner{
		accountsSrv:    accountsSrv,
		docSrv:         docSrv,
		dispatcher:     dispatcher,
		api:            uniquesAPI,
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
