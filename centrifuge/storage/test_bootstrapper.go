// +build integration unit

package storage

import (
	"fmt"

	"github.com/centrifuge/go-centrifuge/centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/centrifuge/config"
	"github.com/centrifuge/go-centrifuge/centrifuge/utils"
)

const testStoragePath = "/tmp/centrifuge_data.leveldb_TESTING"

func (*Bootstrapper) TestBootstrap(context map[string]interface{}) error {
	rs := getRandomTestStoragePath()
	config.Config.V.SetDefault("storage.Path", rs)
	log.Info("Set storage.Path to:", config.Config.GetStoragePath())
	levelDB, err := NewLevelDBStorage(config.Config.GetStoragePath())
	if err != nil {
		return fmt.Errorf("failed to init level db: %v", err)
	}

	log.Infof("Setting levelDb at: %s", config.Config.GetStoragePath())
	context[bootstrap.BootstrappedLevelDb] = levelDB
	return nil
}

func (*Bootstrapper) TestTearDown() error {
	CloseLevelDBStorage()
	// os.RemoveAll(config.Config.GetStoragePath())
	return nil
}

func getRandomTestStoragePath() string {
	return fmt.Sprintf("%s_%x", testStoragePath, utils.RandomByte32())
}
