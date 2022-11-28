package leveldb

import (
	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/storage"
)

// BootstrappedLevelDB key for bootstrap leveldb
const BootstrappedLevelDB = "BootstrappedLevelDB"

// Bootstrapper implements bootstrapper.Bootstrapper.
type Bootstrapper struct{}

// Bootstrap initialises the levelDB.
func (*Bootstrapper) Bootstrap(context map[string]interface{}) error {
	cfg, ok := context[bootstrap.BootstrappedConfig].(config.Configuration)

	if !ok {
		return errors.New("config not initialised")
	}

	configLevelDB, err := NewLevelDBStorage(cfg.GetConfigStoragePath())
	if err != nil {
		return errors.New("failed to init config level db: %v", err)
	}
	context[storage.BootstrappedConfigDB] = NewLevelDBRepository(configLevelDB)

	levelDB, err := NewLevelDBStorage(cfg.GetStoragePath())
	if err != nil {
		return errors.New("failed to init level db: %v", err)
	}
	context[storage.BootstrappedDB] = NewLevelDBRepository(levelDB)
	context[BootstrappedLevelDB] = levelDB
	return nil
}
