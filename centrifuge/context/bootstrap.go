package context

import (
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/config"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/coredocument/repository"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/ethereum"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/invoice/repository"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/purchaseorder/repository"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/signatures"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/storage"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/version"
	logging "github.com/ipfs/go-log"
)

var log = logging.Logger("context")

func Bootstrap() {
	log.Infof("Running cent node on version: %s", version.GetVersion())
	config.Config.InitializeViper()

	levelDB := storage.NewLevelDBStorage(config.Config.GetStoragePath())
	coredocumentrepository.NewLevelDBRepository(&coredocumentrepository.LevelDBRepository{LevelDB: levelDB})
	invoicerepository.NewLevelDBInvoiceRepository(&invoicerepository.LevelDBInvoiceRepository{Leveldb: levelDB})
	purchaseorderrepository.InitLevelDBRepository(levelDB)
	signatures.NewSigningService(signatures.SigningService{})
	createEthereumConnection()
}

func createEthereumConnection() {
	client := ethereum.NewClientConnection()
	ethereum.SetConnection(client)
}
