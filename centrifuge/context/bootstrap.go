package context

import (
	"fmt"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/config"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/coredocument/repository"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/ethereum"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/invoice/repository"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/purchaseorder/repository"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/storage"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/tools"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/version"
	logging "github.com/ipfs/go-log"
	gologging "github.com/whyrusleeping/go-logging"
	"os"
	"path/filepath"
	"strings"
)

const testStoragePath = "/tmp/centrifuge_data.leveldb_TESTING"

var log = logging.Logger("context")

func Bootstrap() {
	log.Infof("Running cent node on version: %s", version.GetVersion())
	config.Config.InitializeViper()

	levelDB := storage.NewLeveldbStorage(config.Config.GetStoragePath())
	coredocumentrepository.NewLevelDBCoreDocumentRepository(&coredocumentrepository.LevelDBCoreDocumentRepository{levelDB})
	invoicerepository.NewLevelDBInvoiceRepository(&invoicerepository.LevelDBInvoiceRepository{levelDB})
	purchaseorderrepository.NewLevelDBPurchaseOrderRepository(&purchaseorderrepository.LevelDBPurchaseOrderRepository{levelDB})
	createEthereumConnection()
}

func createEthereumConnection() {
	client := ethereum.NewClientConnection()
	ethereum.SetConnection(client)
}

func getRandomTestStoragePath() string {
	return fmt.Sprintf("%s_%x", testStoragePath, tools.RandomByte32())
}

func TestBootstrap() {
	TestUnitBootstrap()
	createEthereumConnection()
}

func TestUnitBootstrap() {
	logging.SetAllLoggers(gologging.DEBUG)
	backend := gologging.NewLogBackend(os.Stdout, "", 0)
	gologging.SetBackend(backend)

	// To get the config location, we need to traverse the path to find the `go-centrifuge` folder
	path, _ := filepath.Abs("./")
	match := ""
	for match == "" {
		path = filepath.Join(path, "../")
		if strings.HasSuffix(path, "go-centrifuge") {
			match = path
		}
		if filepath.Dir(path) == "/" {
			log.Fatal("Current working dir is not in `go-centrifuge`")
		}
	}
	config.Config = config.NewConfiguration(fmt.Sprintf("%s/resources/testing_config.yaml", match))
	config.Config.InitializeViper()
	rs := getRandomTestStoragePath()
	config.Config.V.SetDefault("storage.Path", rs)
	log.Info("Set storage.Path to:", config.Config.GetStoragePath())
	config.Config.V.WriteConfigAs("/tmp/cent_config.yaml")

	log.Infof("Creating levelDb at: %s", config.Config.GetStoragePath())
	levelDB := storage.NewLeveldbStorage(config.Config.GetStoragePath())
	coredocumentrepository.NewLevelDBCoreDocumentRepository(&coredocumentrepository.LevelDBCoreDocumentRepository{levelDB})
	invoicerepository.NewLevelDBInvoiceRepository(&invoicerepository.LevelDBInvoiceRepository{levelDB})
	purchaseorderrepository.NewLevelDBPurchaseOrderRepository(&purchaseorderrepository.LevelDBPurchaseOrderRepository{levelDB})
}

func TestTearDown() {
	storage.CloseLeveldbStorage()
	os.RemoveAll(config.Config.GetStoragePath())
	config.Config = nil
}
