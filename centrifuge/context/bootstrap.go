package context

import (
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/bootstrapper"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/config"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/coredocument/repository"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/ethereum"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/invoice/repository"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/purchaseorder/repository"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/signatures"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/storage"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/version"
	"github.com/ethereum/go-ethereum/log"
)

func Bootstrap() {
	context := make(map[string]interface{})
	for _, b := range []bootstrapper.Bootstrapper{
		&version.Bootstrapper{},
		&config.Bootstrapper{},
		&storage.Bootstrapper{},
		&coredocumentrepository.Bootstrapper{},
		&invoicerepository.Bootstrapper{},
		&purchaseorderrepository.Bootstrapper{},
		&signatures.Bootstrapper{},
		&ethereum.Bootstrapper{},
	} {
		err := b.Bootstrap(context)
		if err != nil {
			log.Error("Error encountered while bootstrapping", err)
			panic("Terminating node on bootstrap error")
		}
	}
}
