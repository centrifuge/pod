package storage

import (
	"errors"
	"fmt"

	"github.com/centrifuge/go-centrifuge/config"
)

const BootstrappedLevelDb string = "BootstrappedLevelDb"

type Bootstrapper struct {
}

func (*Bootstrapper) Bootstrap(context map[string]interface{}) error {
	if _, ok := context[config.BootstrappedConfig]; !ok {
		return errors.New("config not initialised")
	}
	cfg := context[config.BootstrappedConfig].(*config.Configuration)

	levelDB, err := NewLevelDBStorage(cfg.GetStoragePath())
	if err != nil {
		return fmt.Errorf("failed to init level db: %v", err)
	}

	context[BootstrappedLevelDb] = levelDB
	return nil
}
