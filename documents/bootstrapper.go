package documents

import (
	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/queue"
	"github.com/centrifuge/go-centrifuge/storage"
	"github.com/centrifuge/go-centrifuge/transactions"
)

const (
	// BootstrappedRegistry is the key to ServiceRegistry in Bootstrap context
	BootstrappedRegistry = "BootstrappedRegistry"

	// BootstrappedDocumentRepository is the key to the database repository of documents
	BootstrappedDocumentRepository = "BootstrappedDocumentRepository"
)

// Bootstrapper implements bootstrap.Bootstrapper.
type Bootstrapper struct{}

// Bootstrap sets the required storage and registers
func (Bootstrapper) Bootstrap(ctx map[string]interface{}) error {
	ctx[BootstrappedRegistry] = NewServiceRegistry()
	ldb, ok := ctx[storage.BootstrappedDB].(storage.Repository)
	if !ok {
		return ErrDocumentBootstrap
	}
	ctx[BootstrappedDocumentRepository] = NewDBRepository(ldb)
	return nil
}

type PostBootstrapper struct{}

func (PostBootstrapper) Bootstrap(ctx map[string]interface{}) error {
	queueSrv, ok := ctx[bootstrap.BootstrappedQueueServer].(*queue.Server)
	if !ok {
		return errors.New("queue not initialised")
	}

	coreDocProc, ok := ctx[bootstrap.BootstrappedCoreDocProc].(anchorProcessor)
	if !ok {
		return errors.New("coredoc processor not initialised")
	}

	repo, ok := ctx[BootstrappedDocumentRepository].(Repository)
	if !ok {
		return errors.New("document repository not initialised")
	}

	task := &documentAnchorTask{
		BaseTask: transactions.BaseTask{
			TxRepository: ctx[transactions.BootstrappedRepo].(transactions.Repository),
		},
		config:        ctx[bootstrap.BootstrappedConfig].(config.Configuration),
		processor:     coreDocProc,
		modelGetFunc:  repo.Get,
		modelSaveFunc: repo.Update,
	}
	queueSrv.RegisterTaskType(documentAnchorTaskName, task)
	return nil
}
