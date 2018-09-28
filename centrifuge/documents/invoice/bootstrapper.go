package invoice

import (
	"errors"

	"github.com/centrifuge/centrifuge-protobufs/documenttypes"
	"github.com/centrifuge/go-centrifuge/centrifuge/documents"

	"github.com/centrifuge/go-centrifuge/centrifuge/bootstrapper"
	"github.com/centrifuge/go-centrifuge/centrifuge/storage"
)

type Bootstrapper struct{}

func (b *Bootstrapper) Bootstrap(context map[string]interface{}) error {
	if _, ok := context[bootstrapper.BootstrappedLevelDb]; ok {
		InitLevelDBRepository(storage.GetLevelDBStorage())

		err := b.registerInvoiceService()
		return err
	}

	return errors.New("could not initialize invoice repository")
}

func (*Bootstrapper) registerInvoiceService() error {

	return documents.GetRegistryInstance().Register(documenttypes.InvoiceDataTypeUrl, &service{})
}
