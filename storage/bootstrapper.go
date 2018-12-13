package storage

import (
	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/errors"
)

const (
	// BootstrappedLevelDB is a key mapped to levelDB at boot
	BootstrappedLevelDB string = "BootstrappedLevelDB"
	// BootstrappedConfigLevelDB is a key mapped to levelDB for configs at boot
	BootstrappedConfigLevelDB string = "BootstrappedConfigLevelDB"
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
	context[BootstrappedConfigLevelDB] = configLevelDB

	levelDB, err := NewLevelDBStorage(cfg.GetStoragePath())
	if err != nil {
		return errors.New("failed to init level db: %v", err)
	}
	context[BootstrappedLevelDB] = levelDB
	return nil
}
