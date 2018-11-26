// +build integration unit

package storage

import (
	"fmt"

	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/utils"
)

const testStoragePath = "/tmp/centrifuge_data.leveldb_TESTING"

func (*Bootstrapper) TestBootstrap(context map[string]interface{}) error {
	rs := getRandomTestStoragePath()
	cfg := context[config.BootstrappedConfig].(*config.Configuration)
	cfg.SetDefault("storage.Path", rs)
	log.Info("Set storage.Path to:", cfg.GetStoragePath())
	levelDB, err := NewLevelDBStorage(cfg.GetStoragePath())
	if err != nil {
		return fmt.Errorf("failed to init level db: %v", err)
	}

	log.Infof("Setting levelDb at: %s", cfg.GetStoragePath())
	context[BootstrappedLevelDB] = levelDB
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
