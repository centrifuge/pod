package testing

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/CentrifugeInc/go-centrifuge/centrifuge/config"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/ethereum"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/storage"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/tools"
	logging "github.com/ipfs/go-log"
	"github.com/syndtr/goleveldb/leveldb"
	gologging "github.com/whyrusleeping/go-logging"
)

const testStoragePath = "/tmp/centrifuge_data.leveldb_TESTING"

var log = logging.Logger("context")

func createEthereumConnection() {
	client := ethereum.NewClientConnection()
	ethereum.SetConnection(client)
}

func getRandomTestStoragePath() string {
	return fmt.Sprintf("%s_%x", testStoragePath, tools.RandomByte32())
}

func GetLevelDBStorage() *leveldb.DB {
	return storage.NewLevelDBStorage(config.Config.GetStoragePath())
}

// ---- Ethereum ----
func TestFunctionalEthereumBootstrap() {
	TestIntegrationBootstrap()
	createEthereumConnection()
}
func TestFunctionalEthereumTearDown() {
	TestIntegrationTearDown()
}

// ---- END Ethereum ----

// ---- Integration Testing ----
func TestIntegrationBootstrap() {
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
}

func TestIntegrationTearDown() {
	storage.CloseLevelDBStorage()
	os.RemoveAll(config.Config.GetStoragePath())
	config.Config = nil
}

// ---- End Integration Testing ----
