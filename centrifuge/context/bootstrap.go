package context

import (
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/config"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/coredocument/repository"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/ethereum"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/identity"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/invoice/repository"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/purchaseorder/repository"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/queue"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/signatures"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/storage"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/version"
	logging "github.com/ipfs/go-log"
)

var log = logging.Logger("context")

func Bootstrap() {
	log.Infof("Running cent node on version: %s", version.GetVersion())
	config.Config.InitializeViper()

	levelDB := storage.NewLeveldbStorage(config.Config.GetStoragePath())
	coredocumentrepository.NewLevelDBCoreDocumentRepository(&coredocumentrepository.LevelDBCoreDocumentRepository{levelDB})
	invoicerepository.NewLevelDBInvoiceRepository(&invoicerepository.LevelDBInvoiceRepository{levelDB})
	purchaseorderrepository.NewLevelDBPurchaseOrderRepository(&purchaseorderrepository.LevelDBPurchaseOrderRepository{levelDB})
	signatures.NewSigningService(signatures.SigningService{})
	createEthereumConnection()
	bootstrapQueuing()
}

func bootstrapQueuing() {
	queue.InitQueue([]queue.QueuedTask{
		&identity.IdRegistrationConfirmationTask{},
	})
}

func createEthereumConnection() {
	client := ethereum.NewClientConnection()
	ethereum.SetConnection(client)
}
