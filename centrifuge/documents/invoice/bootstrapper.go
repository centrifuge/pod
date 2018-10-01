package invoice

import (
	"errors"

	"github.com/centrifuge/go-centrifuge/centrifuge/bootstrap"
	"github.com/centrifuge/centrifuge-protobufs/documenttypes"
	"github.com/centrifuge/go-centrifuge/centrifuge/documents"

	"github.com/centrifuge/go-centrifuge/centrifuge/storage"
)

type Bootstrapper struct{}

func (*Bootstrapper) Bootstrap(context map[string]interface{}) error {
	if _, ok := context[bootstrap.BootstrappedLevelDb]; ok {
		InitLevelDBRepository(storage.GetLevelDBStorage())

	} else {
		return errors.New("could not initialize invoice repository")
	}

	return registerInvoiceService()

}

func registerInvoiceService() error {

	return documents.GetRegistryInstance().Register(documenttypes.InvoiceDataTypeUrl, &service{})
}
