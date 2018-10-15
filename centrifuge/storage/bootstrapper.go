package storage

import (
	"errors"

	"github.com/centrifuge/go-centrifuge/centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/centrifuge/config"
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

func (b *Bootstrapper) TestBootstrap(context map[string]interface{}) error {
	return b.Bootstrap(context)
}
