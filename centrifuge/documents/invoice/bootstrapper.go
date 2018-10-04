package invoice

import (
	"errors"

	"github.com/centrifuge/centrifuge-protobufs/documenttypes"
	"github.com/centrifuge/go-centrifuge/centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/centrifuge/storage"
)

type Bootstrapper struct{}

// Bootstrap sets the required storage and registers
func (*Bootstrapper) Bootstrap(context map[string]interface{}) error {
	if _, ok := context[bootstrap.BootstrappedLevelDb]; !ok {
		return errors.New("initializing LevelDB repository failed")
	}

	InitLegacyRepository(storage.GetLevelDBStorage())
	return registerInvoiceService()
}

func registerInvoiceService() error {
	return documents.GetRegistryInstance().Register(documenttypes.InvoiceDataTypeUrl, &service{})
}
