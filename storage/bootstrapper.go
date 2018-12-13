package storage

import (
	"github.com/centrifuge/go-centrifuge/errors"

	"github.com/centrifuge/go-centrifuge/config"
)

// BootstrappedLevelDB is a key mapped to levelDB in the boot
const BootstrappedLevelDB string = "BootstrappedLevelDB"

// Bootstrapper implements bootstrapper.Bootstrapper.
type Bootstrapper struct{}

// Bootstrap initialises the levelDB.
func (*Bootstrapper) Bootstrap(context map[string]interface{}) error {
	if _, ok := context[config.BootstrappedConfig]; !ok {
		return errors.New("config not initialised")
	}
	cfg := context[config.BootstrappedConfig].(config.Configuration)

	levelDB, err := NewLevelDBStorage(cfg.GetStoragePath())
	if err != nil {
		return errors.New("failed to init level db: %v", err)
	}

	context[BootstrappedLevelDB] = levelDB
	return nil
}
