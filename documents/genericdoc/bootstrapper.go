package genericdoc

import (
	"github.com/centrifuge/go-centrifuge/config/configstore"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/identity/ethid"

	"github.com/centrifuge/go-centrifuge/anchors"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/identity"
)

// Bootstrapper implements bootstrap.Bootstrapper.
type Bootstrapper struct{}

// BootstrappedGenService is a key mapped to the generic document service
const BootstrappedGenService = "BootstrappedGenService "

// Bootstrap sets the required storage and registers
func (Bootstrapper) Bootstrap(ctx map[string]interface{}) error {
	cfg, err := configstore.RetrieveConfig(true, ctx)
	if err != nil {
		return err
	}

	anchorRepo, ok := ctx[anchors.BootstrappedAnchorRepo].(anchors.AnchorRepository)
	if !ok {
		return errors.New("anchor repository not initialised")
	}

	idService, ok := ctx[ethid.BootstrappedIDService].(identity.Service)
	if !ok {
		return errors.New("identity service not initialised")
	}

	repo, ok := ctx[documents.BootstrappedDocumentRepository].(documents.Repository)
	if !ok {
		return errors.New("document db repository not initialised")
	}

	ctx[BootstrappedGenService] = DefaultService(cfg, repo, anchorRepo, idService)
	return nil

}
