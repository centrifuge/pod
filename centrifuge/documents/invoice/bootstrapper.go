package invoice

import (
	"errors"

	"github.com/centrifuge/centrifuge-protobufs/documenttypes"
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

	InitLegacyRepository(storage.GetLevelDBStorage())
	return registerInvoiceService()
}

func registerInvoiceService() error {
	// TODO coredocument processor and IDService usage here looks shitty(unnecessary dependency), needs to change soon
	invoiceService := DefaultService(
		GetRepository(),
		coredocumentprocessor.DefaultProcessor(identity.IDService, p2p.NewP2PClient()))
	return documents.GetRegistryInstance().Register(documenttypes.InvoiceDataTypeUrl, invoiceService)
}
