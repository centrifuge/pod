package leveldb

import (
	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/storage"
)

// Config holds configuration data for storage package
type Config interface {
	GetStoragePath() string
	GetConfigStoragePath() string
	SetDefault(key string, value interface{})
}

// Bootstrapper implements bootstrapper.Bootstrapper.
type Bootstrapper struct{}

// Bootstrap initialises the levelDB.
func (*Bootstrapper) Bootstrap(context map[string]interface{}) error {
	if _, ok := context[bootstrap.BootstrappedConfig]; !ok {
		return errors.New("config not initialised")
	}
	cfg := context[bootstrap.BootstrappedConfig].(Config)

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
	return nil
}
