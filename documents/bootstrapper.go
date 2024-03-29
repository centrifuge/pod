package documents

import (
	"github.com/centrifuge/pod/bootstrap"
	"github.com/centrifuge/pod/config"
	"github.com/centrifuge/pod/errors"
	v2 "github.com/centrifuge/pod/identity/v2"
	"github.com/centrifuge/pod/jobs"
	"github.com/centrifuge/pod/notification"
	"github.com/centrifuge/pod/pallets"
	"github.com/centrifuge/pod/pallets/anchors"
	"github.com/centrifuge/pod/storage"
)

const (
	// BootstrappedRegistry is the key to ServiceRegistry in Bootstrap context
	BootstrappedRegistry = "BootstrappedRegistry"

	// BootstrappedDocumentRepository is the key to the database repository of documents
	BootstrappedDocumentRepository = "BootstrappedDocumentRepository"

	// BootstrappedDocumentService is the key to bootstrapped document service
	BootstrappedDocumentService = "BootstrappedDocumentService"

	// BootstrappedAnchorProcessor is the key to bootstrapped anchor processor
	BootstrappedAnchorProcessor = "BootstrappedAnchorProcessor"
)

// Bootstrapper implements bootstrap.Bootstrapper.
type Bootstrapper struct{}

// Bootstrap sets the required storage and registers
func (Bootstrapper) Bootstrap(ctx map[string]interface{}) error {
	registry := NewServiceRegistry()

	ldb, ok := ctx[storage.BootstrappedDB].(storage.Repository)
	if !ok {
		return ErrDocumentBootstrap
	}

	repo := NewDBRepository(ldb)
	anchorSrv, ok := ctx[pallets.BootstrappedAnchorService].(anchors.API)
	if !ok {
		return errors.New("anchor repository not initialised")
	}

	dispatcher, ok := ctx[jobs.BootstrappedJobDispatcher].(jobs.Dispatcher)
	if !ok {
		return errors.New("jobs dispatcher not initialised")
	}

	identityService, ok := ctx[v2.BootstrappedIdentityServiceV2].(v2.Service)
	if !ok {
		return errors.New("identity service not initialised")
	}

	notifier := notification.NewWebhookSender()

	ctx[BootstrappedDocumentService] = NewService(
		repo,
		anchorSrv,
		registry,
		dispatcher,
		identityService,
		notifier,
	)

	ctx[BootstrappedRegistry] = registry
	ctx[BootstrappedDocumentRepository] = repo
	return nil
}

// PostBootstrapper to run the post after all bootstrappers.
type PostBootstrapper struct{}

// Bootstrap register task to the queue.
func (PostBootstrapper) Bootstrap(ctx map[string]interface{}) error {
	cfgService, ok := ctx[config.BootstrappedConfigStorage].(config.Service)
	if !ok {
		return errors.New("config service not initialised")
	}

	repo, ok := ctx[BootstrappedDocumentRepository].(Repository)
	if !ok {
		return errors.New("document repository not initialised")
	}

	anchorSrv, ok := ctx[pallets.BootstrappedAnchorService].(anchors.API)
	if !ok {
		return errors.New("anchor repository not initialised")
	}

	cfg, ok := ctx[bootstrap.BootstrappedConfig].(config.Configuration)
	if !ok {
		return errors.New("documents config not initialised")
	}

	p2pClient, ok := ctx[bootstrap.BootstrappedPeer].(Client)
	if !ok {
		return errors.New("p2p client not initialised")
	}

	identityService, ok := ctx[v2.BootstrappedIdentityServiceV2].(v2.Service)

	if !ok {
		return errors.New("identity service v2 not initialised")
	}

	dp := NewAnchorProcessor(p2pClient, anchorSrv, cfg, identityService)
	ctx[BootstrappedAnchorProcessor] = dp

	dispatcher := ctx[jobs.BootstrappedJobDispatcher].(jobs.Dispatcher)

	go dispatcher.RegisterRunner(anchorJob, &AnchorJob{
		configSrv: cfgService,
		processor: dp,
		repo:      repo,
	})

	return nil
}
