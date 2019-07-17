package generic

import (
	"github.com/centrifuge/centrifuge-protobufs/documenttypes"
	"github.com/centrifuge/go-centrifuge/anchors"
	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/jobs"
	"github.com/centrifuge/go-centrifuge/queue"
)

// Bootstrapper implements bootstrap.Bootstrapper.
type Bootstrapper struct{}

// Bootstrap sets the required storage and registers
func (Bootstrapper) Bootstrap(ctx map[string]interface{}) error {
	registry, ok := ctx[documents.BootstrappedRegistry].(*documents.ServiceRegistry)
	if !ok {
		return errors.New("service registry not initialised")
	}

	docSrv, ok := ctx[documents.BootstrappedDocumentService].(documents.Service)
	if !ok {
		return errors.New("document service not initialised")
	}

	repo, ok := ctx[documents.BootstrappedDocumentRepository].(documents.Repository)
	if !ok {
		return errors.New("document db repository not initialised")
	}
	repo.Register(&Generic{})

	queueSrv, ok := ctx[bootstrap.BootstrappedQueueServer].(*queue.Server)
	if !ok {
		return errors.New("queue server not initialised")
	}

	jobManager, ok := ctx[jobs.BootstrappedService].(jobs.Manager)
	if !ok {
		return errors.New("transaction service not initialised")
	}

	anchorRepo, ok := ctx[anchors.BootstrappedAnchorRepo].(anchors.AnchorRepository)
	if !ok {
		return anchors.ErrAnchorRepoNotInitialised
	}

	// register service
	srv := DefaultService(docSrv, repo, queueSrv, jobManager, anchorRepo)
	err := registry.Register(documenttypes.GenericDataTypeUrl, srv)
	if err != nil {
		return errors.New("failed to register generic doc service: %v", err)
	}

	err = registry.Register(Scheme, srv)
	if err != nil {
		return errors.New("failed to register generic doc service: %v", err)
	}

	return nil
}
