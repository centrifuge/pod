// +build integration unit

package storage

import (
	"fmt"

	"github.com/centrifuge/go-centrifuge/errors"

	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/syndtr/goleveldb/leveldb"
)

var db *leveldb.DB
var configdb *leveldb.DB

func (*Bootstrapper) TestBootstrap(context map[string]interface{}) (err error) {
	cfg := context[bootstrap.BootstrappedConfig].(Config)

	crs := GetRandomTestStoragePath()
	cfg.SetDefault("configStorage.path", crs)
	log.Info("Set configStorage.path to:", cfg.GetConfigStoragePath())
	configdb, err = NewLevelDBStorage(cfg.GetConfigStoragePath())
	if err != nil {
		return fmt.Errorf("failed to init config level db: %v", err)
	}
	context[BootstrappedConfigLevelDB] = configdb

	rs := GetRandomTestStoragePath()
	cfg.SetDefault("storage.Path", rs)
	log.Info("Set storage.Path to:", cfg.GetStoragePath())
	db, err = NewLevelDBStorage(cfg.GetStoragePath())
	if err != nil {
		return fmt.Errorf("failed to init level db: %v", err)
	}
	log.Infof("Setting levelDb at: %s", cfg.GetStoragePath())
	context[BootstrappedLevelDB] = db
	return nil
}

func (b *Bootstrapper) TestTearDown() error {
	var err error
	dbs := []*leveldb.DB{db, configdb}
	for _, idb := range dbs {
		if ierr := idb.Close(); ierr != nil {
			if err == nil {
				err = errors.New("%s", ierr)
			} else {
				err = errors.AppendError(err, ierr)
			}
		}
	}
	return err
}
