package context

import (
	"github.com/centrifuge/go-centrifuge/centrifuge/anchors"
	"github.com/centrifuge/go-centrifuge/centrifuge/bootstrapper"
	"github.com/centrifuge/go-centrifuge/centrifuge/config"
	"github.com/centrifuge/go-centrifuge/centrifuge/coredocument/repository"
	"github.com/centrifuge/go-centrifuge/centrifuge/ethereum"
	"github.com/centrifuge/go-centrifuge/centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/centrifuge/invoice/repository"
	"github.com/centrifuge/go-centrifuge/centrifuge/purchaseorder/repository"
	"github.com/centrifuge/go-centrifuge/centrifuge/queue"
	"github.com/centrifuge/go-centrifuge/centrifuge/storage"
	"github.com/centrifuge/go-centrifuge/centrifuge/version"
	logging "github.com/ipfs/go-log"
)

var log = logging.Logger("context")

type MainBootstrapper struct {
	Bootstrappers []bootstrapper.Bootstrapper
}

func (m *MainBootstrapper) PopulateDefaultBootstrappers() {
	m.Bootstrappers = []bootstrapper.Bootstrapper{
		&version.Bootstrapper{},
		&config.Bootstrapper{},
		&storage.Bootstrapper{},
		&coredocumentrepository.Bootstrapper{},
		&invoicerepository.Bootstrapper{},
		&purchaseorderrepository.Bootstrapper{},
		&ethereum.Bootstrapper{},
		&anchors.Bootstrapper{},
		&identity.Bootstrapper{},
		&queue.Bootstrapper{},
	}
}

func (m *MainBootstrapper) Bootstrap(context map[string]interface{}) error {
	for _, b := range m.Bootstrappers {
		err := b.Bootstrap(context)
		if err != nil {
			log.Error("Error encountered while bootstrapping", err)
			return err
		}
	}
	return nil
}
