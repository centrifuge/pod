package context

import (
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/config"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/coredocument/repository"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/invoice/repository"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/storage"
	logging "github.com/ipfs/go-log"
	gologging "github.com/whyrusleeping/go-logging"
	"os"
)

const TestStoragePath = "/tmp/centrifuge_data.leveldb_TESTING"

var log = logging.Logger("context")
var format = gologging.MustStringFormatter(
	"%{color}%{time:15:04:05.000} %{shortfunc} ▶ %{level:.4s} %{id:03x}%{color:reset} %{message}",
)

func Bootstrap() {
	config.Config.InitializeViper()

	levelDB := storage.NewLeveldbStorage(config.Config.GetStoragePath())

	coredocumentrepository.NewLevelDBCoreDocumentRepository(&coredocumentrepository.LevelDBCoreDocumentRepository{levelDB})
	invoicerepository.NewLevelDBInvoiceRepository(&invoicerepository.LevelDBInvoiceRepository{levelDB})
}

func TestBootstrap() {
	logging.SetAllLoggers(gologging.INFO) // Change to DEBUG for extra info
	backend := gologging.NewLogBackend(os.Stdout, "", 0)
	gologging.SetBackend(backend)

	if config.Config.V == nil {
		err := config.Config.SetConfigFile("../../resources/testing_config.yaml")
		if err != nil {
			panic(err)
		}
		config.Config.InitializeViper()
		config.Config.V.SetDefault("storage.Path", TestStoragePath)
	}

	config.Config.V.Set("centrifugeNetwork", "testing")
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
