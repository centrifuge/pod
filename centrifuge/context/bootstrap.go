package context

import (
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/config"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/coredocument/repository"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/invoice/repository"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/storage"
	logging "github.com/ipfs/go-log"
	"os"
)

var log = logging.Logger("bootstrap")

const TestStoragePath = "/tmp/centrifuge_data.leveldb_TESTING"

func Bootstrap() {
	config.Config.InitializeViper()

	levelDB := storage.NewLeveldbStorage(config.Config.GetStoragePath())

	coredocumentrepository.NewLevelDBCoreDocumentRepository(&coredocumentrepository.LevelDBCoreDocumentRepository{levelDB})
	invoicerepository.NewLevelDBInvoiceRepository(&invoicerepository.LevelDBInvoiceRepository{levelDB})
}

func TestBootstrap() {
	if config.Config.V == nil {
		err := config.Config.SetConfigFile("../../resources/testing_config.yaml")
		if err != nil {
			panic(err)
		}
		config.Config.InitializeViper()
		config.Config.V.SetDefault("storage.Path", TestStoragePath)
	}

	config.Config.V.Set("centrifugeNetwork", "testing")
	log.Info("Writing config to /tmp/cent_config.yaml")
	config.Config.V.WriteConfigAs("/tmp/cent_config.yaml")

	levelDB := storage.NewLeveldbStorage(config.Config.GetStoragePath())

	coredocumentrepository.NewLevelDBCoreDocumentRepository(&coredocumentrepository.LevelDBCoreDocumentRepository{levelDB})
	invoicerepository.NewLevelDBInvoiceRepository(&invoicerepository.LevelDBInvoiceRepository{levelDB})
}

func TestTearDown() {
	Close()
	if config.Config.GetStoragePath() == TestStoragePath {
		os.RemoveAll(TestStoragePath)
	}
}

func Close() {
	levelDB := storage.NewLeveldbStorage(config.Config.GetStoragePath())
	levelDB.Close()
}
