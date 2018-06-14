package context

import (
	"fmt"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/config"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/coredocument/repository"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/invoice/repository"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/storage"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/tools"
	logging "github.com/ipfs/go-log"
	gologging "github.com/whyrusleeping/go-logging"
	"os"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/ethereum"
	"github.com/ethereum/go-ethereum/ethclient"
	"sync"
	"net/url"
	"github.com/ethereum/go-ethereum/rpc"
)

const testStoragePath = "/tmp/centrifuge_data.leveldb_TESTING"

var log = logging.Logger("context")

func Bootstrap() {
	config.Config.InitializeViper()

	levelDB := storage.NewLeveldbStorage(config.Config.GetStoragePath())
	coredocumentrepository.NewLevelDBCoreDocumentRepository(&coredocumentrepository.LevelDBCoreDocumentRepository{levelDB})
	invoicerepository.NewLevelDBInvoiceRepository(&invoicerepository.LevelDBInvoiceRepository{levelDB})
	createEthereumConnection()
}

func createEthereumConnection() {
	log.Info("Opening connection to Ethereum:", config.Config.GetEthereumNodeURL())
	u, err := url.Parse(config.Config.GetEthereumNodeURL())
	if err != nil {
		log.Fatalf("Failed to connect to parse ethereum.gethSocket URL: %v", err)
	}
	c, err := rpc.Dial(u.String())
	if err != nil {
		log.Fatalf("Failed to connect to the Ethereum client [%s]: %v", u.String(), err)
	}
	client := ethclient.NewClient(c)
	if err != nil {
		log.Fatalf("Failed to connect to the Ethereum client [%s]: %v", u.String(), err)
	} else {
		ethereum.SetConnection(ethereum.GethClient{client,c, u, &sync.Mutex{}})
	}
}

func getRandomTestStoragePath() string {
	return fmt.Sprintf("%s_%x", testStoragePath, tools.RandomByte32())
}

func TestBootstrap() {
	logging.SetAllLoggers(gologging.DEBUG)
	backend := gologging.NewLogBackend(os.Stdout, "", 0)
	gologging.SetBackend(backend)

	projectPath, _ := os.LookupEnv("GOPATH")
	config.Config = config.NewConfiguration(fmt.Sprintf("%s/src/github.com/CentrifugeInc/go-centrifuge/resources/testing_config.yaml", projectPath))
	config.Config.InitializeViper()
	rs := getRandomTestStoragePath()
	config.Config.V.SetDefault("storage.Path", rs)
	log.Info("Set storage.Path to:", config.Config.GetStoragePath())
	config.Config.V.WriteConfigAs("/tmp/cent_config.yaml")

	log.Infof("Creating levelDb at: %s", config.Config.GetStoragePath())
	levelDB := storage.NewLeveldbStorage(config.Config.GetStoragePath())
	coredocumentrepository.NewLevelDBCoreDocumentRepository(&coredocumentrepository.LevelDBCoreDocumentRepository{levelDB})
	invoicerepository.NewLevelDBInvoiceRepository(&invoicerepository.LevelDBInvoiceRepository{levelDB})
	createEthereumConnection()
}

func TestTearDown() {
	storage.CloseLeveldbStorage()
	os.RemoveAll(config.Config.GetStoragePath())
	config.Config = nil
}
