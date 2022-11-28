//go:build unit || integration || testworld

package leveldb

import (
	"errors"
	"os"

	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/config"
	centErrors "github.com/centrifuge/go-centrifuge/errors"
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

	log.Debugf("Using config storage path - %s", cfg.GetConfigStoragePath())

	configdb, err = NewLevelDBStorage(cfg.GetConfigStoragePath())
	if err != nil {
		return centErrors.New("failed to init config level db: %v", err)
	}

	context[storage.BootstrappedConfigDB] = NewLevelDBRepository(configdb)

	log.Debugf("Using storage path - %s", cfg.GetStoragePath())

	db, err = NewLevelDBStorage(cfg.GetStoragePath())
	if err != nil {
		return centErrors.New("failed to init level db: %v", err)
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
		closeErr := idb.Close()

		if closeErr != nil && !errors.Is(closeErr, leveldb.ErrClosed) {
			if err == nil {
				err = centErrors.New("%s", closeErr)
			} else {
				err = centErrors.AppendError(err, closeErr)
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
