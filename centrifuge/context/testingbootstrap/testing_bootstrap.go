package testingbootstrap

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/CentrifugeInc/go-centrifuge/centrifuge/anchoring"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/bootstrapper"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/config"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/ethereum"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/identity"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/queue"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/storage"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/tools"
	logging "github.com/ipfs/go-log"
	"github.com/syndtr/goleveldb/leveldb"
	gologging "github.com/whyrusleeping/go-logging"
)

const testStoragePath = "/tmp/centrifuge_data.leveldb_TESTING"

var log = logging.Logger("context")

// ---- Ethereum ----
func TestFunctionalEthereumBootstrap() {
	TestIntegrationBootstrap()
	createEthereumConnection()
	bootstrapQueuing()
}
func TestFunctionalEthereumTearDown() {
	TestIntegrationTearDown()
	tearDownQueuing()
}

// ---- END Ethereum ----

// ---- Integration Testing ----
func TestIntegrationBootstrap() {
	logging.SetAllLoggers(gologging.DEBUG)
	backend := gologging.NewLogBackend(os.Stdout, "", 0)
	gologging.SetBackend(backend)

	InitTestConfig()
	InitTestStoragePath()
	config.Config.V.WriteConfigAs("/tmp/cent_config.yaml")

	log.Infof("Creating levelDb at: %s", config.Config.GetStoragePath())
}

func TestIntegrationTearDown() {
	storage.CloseLevelDBStorage()
	os.RemoveAll(config.Config.GetStoragePath())
	config.Config = nil
}

// ---- End Integration Testing ----

func createEthereumConnection() {
	client, err := ethereum.NewClientConnection()
	if err != nil {
		panic(err)
	}
	ethereum.SetConnection(client)
}

func bootstrapQueuing() {
	// TODO here we would not have to put the bootstrapper.BootstrappedConfig after the TestBootstrapper refactoring
	context := map[string]interface{}{bootstrapper.BootstrappedConfig: true}
	for _, b := range []bootstrapper.TestBootstrapper{
		&anchoring.Bootstrapper{},
		&identity.Bootstrapper{},
		&queue.Bootstrapper{},
	} {
		err := b.TestBootstrap(context)
		if err != nil {
			log.Error("Error encountered while bootstrapping", err)
			panic(err)
		}
	}
}

func tearDownQueuing() {
	queue.StopQueue()
}

func getRandomTestStoragePath() string {
	return fmt.Sprintf("%s_%x", testStoragePath, tools.RandomByte32())
}

func GetLevelDBStorage() *leveldb.DB {
	return storage.NewLevelDBStorage(config.Config.GetStoragePath())
}

func InitTestConfig() {
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
}

func InitTestStoragePath() {
	rs := getRandomTestStoragePath()
	config.Config.V.SetDefault("storage.Path", rs)
	log.Info("Set storage.Path to:", config.Config.GetStoragePath())
}
