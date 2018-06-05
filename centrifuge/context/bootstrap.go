package context

import (
	"fmt"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/config"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/coredocument/repository"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/invoice/repository"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/storage"
	logging "github.com/ipfs/go-log"
	gologging "github.com/whyrusleeping/go-logging"
	"math/rand"
	"os"
)

const testStoragePath = "/tmp/centrifuge_data.leveldb_TESTING"

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

func getRandomTestStoragePath() string {
	return fmt.Sprintf("%s_%d", testStoragePath, rand.Int())
}

func TestBootstrap() {
	logging.SetAllLoggers(gologging.DEBUG)
	backend := gologging.NewLogBackend(os.Stdout, "", 0)
	gologging.SetBackend(backend)

	configPath, _ := os.LookupEnv("GOPATH")
	err := config.Config.SetConfigFile(fmt.Sprintf("%s/src/github.com/CentrifugeInc/go-centrifuge/resources/testing_config.yaml", configPath))
	if err != nil {
		panic(err)
	}
	config.Config.InitializeViper()
	config.Config.V.SetDefault("storage.Path", getRandomTestStoragePath())

	config.Config.V.Set("centrifugeNetwork", "testing")
	config.Config.V.WriteConfigAs("/tmp/cent_config.yaml")

	log.Infof("Creating levelDb at: %s", config.Config.GetStoragePath())
	levelDB := storage.NewLeveldbStorage(config.Config.GetStoragePath())
	coredocumentrepository.NewLevelDBCoreDocumentRepository(&coredocumentrepository.LevelDBCoreDocumentRepository{levelDB})
	invoicerepository.NewLevelDBInvoiceRepository(&invoicerepository.LevelDBInvoiceRepository{levelDB})
}

func TestTearDown() {
	storage.CloseLeveldbStorage()
	os.RemoveAll(config.Config.GetStoragePath())
}
