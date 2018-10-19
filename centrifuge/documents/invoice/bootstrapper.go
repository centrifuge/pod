package invoice

import (
	"errors"
	"fmt"

	"github.com/centrifuge/centrifuge-protobufs/documenttypes"
	"github.com/centrifuge/go-centrifuge/centrifuge/anchors"
	"github.com/centrifuge/go-centrifuge/centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/centrifuge/coredocument/processor"
	"github.com/centrifuge/go-centrifuge/centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/centrifuge/p2p"
	"github.com/centrifuge/go-centrifuge/centrifuge/storage"
)

type Bootstrapper struct{}

// Bootstrap sets the required storage and registers
func (*Bootstrapper) Bootstrap(context map[string]interface{}) error {
	if _, ok := context[bootstrap.BootstrappedLevelDb]; !ok {
		return errors.New("initializing LevelDB repository failed")
	}

	// load legacy db
	InitLegacyRepository(storage.GetLevelDBStorage())

	// register service
	srv := DefaultService(getRepository(), coredocumentprocessor.DefaultProcessor(identity.IDService, p2p.NewP2PClient(), anchors.GetAnchorRepository()))
	err := documents.GetRegistryInstance().Register(documenttypes.InvoiceDataTypeUrl, srv)
	if err != nil {
		return fmt.Errorf("failed to register invoice service: %v", err)
	}

	return nil
}
