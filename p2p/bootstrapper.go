package p2p

import (
	"context"
	"sync"

	"github.com/centrifuge/pod/bootstrap"
	"github.com/centrifuge/pod/config"
	"github.com/centrifuge/pod/dispatcher"
	"github.com/centrifuge/pod/documents"
	"github.com/centrifuge/pod/errors"
	v2 "github.com/centrifuge/pod/identity/v2"
	nftv3 "github.com/centrifuge/pod/nft/v3"
	"github.com/centrifuge/pod/p2p/receiver"
	"github.com/centrifuge/pod/pallets"
	"github.com/centrifuge/pod/pallets/keystore"
	"github.com/libp2p/go-libp2p-core/protocol"
)

// Bootstrapper implements Bootstrapper with p2p details
type Bootstrapper struct {
	testPeerWg        sync.WaitGroup
	testPeerCtx       context.Context
	testPeerCtxCancel context.CancelFunc
}

// Bootstrap initiates p2p server and client into context
func (b *Bootstrapper) Bootstrap(ctx map[string]interface{}) error {
	cfg, err := config.RetrieveConfig(true, ctx)
	if err != nil {
		return err
	}

	cfgService, ok := ctx[config.BootstrappedConfigStorage].(config.Service)
	if !ok {
		return errors.New("configstore not initialised")
	}

	docSrv, ok := ctx[documents.BootstrappedDocumentService].(documents.Service)
	if !ok {
		return errors.New("document service not initialised")
	}

	keystoreAPI, ok := ctx[pallets.BootstrappedKeystoreAPI].(keystore.API)
	if !ok {
		return errors.New("keystore API not initialised")
	}

	identityService, ok := ctx[v2.BootstrappedIdentityServiceV2].(v2.Service)
	if !ok {
		return errors.New("identity service v2 not initialised")
	}

	protocolIDDispatcher, ok := ctx[dispatcher.BootstrappedProtocolIDDispatcher].(dispatcher.Dispatcher[protocol.ID])
	if !ok {
		return errors.New("protocol ID dispatcher not initialised")
	}

	nftService, ok := ctx[nftv3.BootstrappedNFTV3Service].(nftv3.Service)
	if !ok {
		return errors.New("nft service not initialised")
	}

	handler := receiver.NewHandler(
		cfg,
		cfgService,
		receiver.HandshakeValidator(cfg.GetNetworkID(), identityService),
		docSrv,
		identityService,
		nftService,
	)

	ctx[bootstrap.BootstrappedPeer] = newPeer(
		cfg,
		cfgService,
		identityService,
		keystoreAPI,
		protocolIDDispatcher,
		handler,
	)
	return nil
}
