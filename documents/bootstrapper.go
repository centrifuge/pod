package documents

import (
	"github.com/centrifuge/go-centrifuge/anchors"
	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/jobs"
	"github.com/centrifuge/go-centrifuge/jobs/jobsv1"
	"github.com/centrifuge/go-centrifuge/queue"
	"github.com/centrifuge/go-centrifuge/storage"
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
	anchorSrv, ok := ctx[anchors.BootstrappedAnchorService].(anchors.Service)
	if !ok {
		return errors.New("anchor repository not initialised")
	}

	didService, ok := ctx[identity.BootstrappedDIDService].(identity.Service)
	if !ok {
		return errors.New("identity service not initialized")
	}

	cfg, ok := ctx[bootstrap.BootstrappedConfig].(Config)
	if !ok {
		return ErrDocumentConfigNotInitialised
	}

	queueSrv, ok := ctx[bootstrap.BootstrappedQueueServer].(*queue.Server)
	if !ok {
		return errors.New("queue server not initialised")
	}

	jobManager, ok := ctx[jobs.BootstrappedService].(jobs.Manager)
	if !ok {
		return errors.New("transaction service not initialised")
	}

	ctx[BootstrappedDocumentService] = DefaultService(cfg, repo, anchorSrv, registry, didService, queueSrv, jobManager)
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

	queueSrv, ok := ctx[bootstrap.BootstrappedQueueServer].(*queue.Server)
	if !ok {
		return errors.New("queue not initialised")
	}

	repo, ok := ctx[BootstrappedDocumentRepository].(Repository)
	if !ok {
		return errors.New("document repository not initialised")
	}

	anchorSrv, ok := ctx[anchors.BootstrappedAnchorService].(anchors.Service)
	if !ok {
		return errors.New("anchor repository not initialised")
	}

	cfg, ok := ctx[bootstrap.BootstrappedConfig].(Config)
	if !ok {
		return errors.New("documents config not initialised")
	}

	p2pClient, ok := ctx[bootstrap.BootstrappedPeer].(Client)
	if !ok {
		return errors.New("p2p client not initialised")
	}

	didService, ok := ctx[identity.BootstrappedDIDService].(identity.Service)
	if !ok {
		return errors.New("identity service not initialized")
	}

	dp := DefaultProcessor(didService, p2pClient, anchorSrv, cfg)
	ctx[BootstrappedAnchorProcessor] = dp

	jobManager := ctx[jobs.BootstrappedService].(jobs.Manager)
	anchorTask := &documentAnchorTask{
		BaseTask: jobsv1.BaseTask{
			JobManager: jobManager,
		},
		config:        cfgService,
		processor:     dp,
		modelGetFunc:  repo.Get,
		modelSaveFunc: repo.Update,
	}

	queueSrv.RegisterTaskType(documentAnchorTaskName, anchorTask)
	return nil
}
