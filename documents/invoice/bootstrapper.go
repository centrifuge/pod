package invoice

import (
	"errors"
	"fmt"

	"github.com/centrifuge/centrifuge-protobufs/documenttypes"
	"github.com/centrifuge/go-centrifuge/anchors"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/coredocument"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/p2p"
	"github.com/centrifuge/go-centrifuge/storage"
)

type Bootstrapper struct{}

// Bootstrap sets the required storage and registers
func (*Bootstrapper) Bootstrap(context map[string]interface{}) error {
	if _, ok := context[config.BootstrappedConfig]; !ok {
		return errors.New("config hasn't been initialized")
	}
	cfg := context[config.BootstrappedConfig].(*config.Configuration)

	if _, ok := context[storage.BootstrappedLevelDb]; !ok {
		return errors.New("initializing LevelDB repository failed")
	}

	// register service
	srv := DefaultService(cfg, getRepository(), coredocument.DefaultProcessor(identity.IDService, p2p.NewP2PClient(cfg), anchors.GetAnchorRepository(), cfg), anchors.GetAnchorRepository())
	err := documents.GetRegistryInstance().Register(documenttypes.InvoiceDataTypeUrl, srv)
	if err != nil {
		return fmt.Errorf("failed to register invoice service: %v", err)
	}

	return nil
}
