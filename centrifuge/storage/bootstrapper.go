package storage

import (
	"errors"
	"fmt"

	"github.com/centrifuge/go-centrifuge/centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/centrifuge/config"
)

type Bootstrapper struct {
}

func (*Bootstrapper) Bootstrap(context map[string]interface{}) error {
	if _, ok := context[bootstrap.BootstrappedConfig]; !ok {
		return errors.New("config not initialised")
	}

	levelDB, err := NewLevelDBStorage(config.Config.GetStoragePath())
	if err != nil {
		return fmt.Errorf("failed to init level db: %v", err)
	}

	context[bootstrap.BootstrappedLevelDb] = levelDB
	return nil
}
