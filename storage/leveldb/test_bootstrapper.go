// +build integration unit

package leveldb

import (
	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/storage"
	"github.com/syndtr/goleveldb/leveldb"
)

var db *leveldb.DB
var configdb *leveldb.DB

func (*Bootstrapper) TestBootstrap(context map[string]interface{}) (err error) {
	cfg := context[bootstrap.BootstrappedConfig].(config.Configuration)

	crs := GetRandomTestStoragePath()
	cfg.Set("configStorage.path", crs)
	log.Info("Set configStorage.path to:", cfg.GetConfigStoragePath())
	configdb, err = NewLevelDBStorage(cfg.GetConfigStoragePath())
	if err != nil {
		return errors.New("failed to init config level db: %v", err)
	}
	context[storage.BootstrappedConfigDB] = NewLevelDBRepository(configdb)

	rs := GetRandomTestStoragePath()
	cfg.Set("storage.Path", rs)
	log.Info("Set storage.Path to:", cfg.GetStoragePath())
	db, err = NewLevelDBStorage(cfg.GetStoragePath())
	if err != nil {
		return errors.New("failed to init level db: %v", err)
	}
	log.Infof("Setting levelDb at: %s", cfg.GetStoragePath())
	context[storage.BootstrappedDB] = NewLevelDBRepository(db)
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
