//go:build integration || unit
// +build integration unit

package leveldb

import (
	"os"

	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/storage"
	"github.com/syndtr/goleveldb/leveldb"
)

var (
	db                *leveldb.DB
	configdb          *leveldb.DB
	configStoragePath string
	storagePath       string
)

func (*Bootstrapper) TestBootstrap(context map[string]interface{}) (err error) {
	cfg := context[bootstrap.BootstrappedConfig].(config.Configuration)

	configStoragePath = GetRandomTestStoragePath()

	log.Debugf("Using config storage path - %s", configStoragePath)

	configdb, err = NewLevelDBStorage(configStoragePath)
	if err != nil {
		return errors.New("failed to init config level db: %v", err)
	}

	context[storage.BootstrappedConfigDB] = NewLevelDBRepository(configdb)

	storagePath = GetRandomTestStoragePath()

	log.Debugf("Using storage path - %s", storagePath)

	db, err = NewLevelDBStorage(storagePath)
	if err != nil {
		return errors.New("failed to init level db: %v", err)
	}
	log.Infof("Setting levelDb at: %s", cfg.GetStoragePath())
	context[storage.BootstrappedDB] = NewLevelDBRepository(db)
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

	if err := os.RemoveAll(configStoragePath); err != nil {
		return err
	}

	if err := os.RemoveAll(storagePath); err != nil {
		return err
	}

	return err
}
