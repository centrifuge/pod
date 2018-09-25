package storage

import (
	"errors"

	"github.com/CentrifugeInc/go-centrifuge/centrifuge/bootstrap"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/config"
)

type Bootstrapper struct {
}

func (*Bootstrapper) Bootstrap(context map[string]interface{}) error {
	if _, ok := context[bootstrap.BootstrappedConfig]; ok {
		levelDB := NewLevelDBStorage(config.Config.GetStoragePath())
		context[bootstrap.BootstrappedLevelDb] = levelDB
		return nil
	}
	return errors.New("could not initialize levelDB")
}
