// +build integration unit

package config

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/centrifuge/go-centrifuge/storage"
	"github.com/syndtr/goleveldb/leveldb"
)

var db *leveldb.DB

func (*Bootstrapper) TestBootstrap(context map[string]interface{}) error {
	// To get the config location, we need to traverse the path to find the `go-centrifuge` folder
	var err error
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
	cfg := LoadConfiguration(fmt.Sprintf("%s/build/configs/testing_config.yaml", match))
	context[BootstrappedConfig] = cfg

	rs := storage.GetRandomTestStoragePath()
	cfg.SetDefault("storage.Path", rs)
	log.Info("Set storage.Path to:", cfg.GetStoragePath())
	db, err = storage.NewLevelDBStorage(cfg.GetStoragePath())
	if err != nil {
		return fmt.Errorf("failed to init level db: %v", err)
	}

	log.Infof("Setting levelDb at: %s", cfg.GetStoragePath())
	context[BootstrappedLevelDB] = db

	return nil
}

func (b *Bootstrapper) TestTearDown() error {
	return db.Close()
}
