package storage

import (
	"errors"

	"github.com/CentrifugeInc/go-centrifuge/centrifuge/bootstrapper"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/config"
)

type Bootstrapper struct {
}

func (*Bootstrapper) Bootstrap(context map[string]interface{}) error {
	if _, ok := context[bootstrapper.BOOTSTRAPPED_CONFIG]; ok {
		levelDB := NewLevelDBStorage(config.Config.GetStoragePath())
		context[bootstrapper.BOOTSTRAPPED_LEVELDB] = levelDB
		return nil
	}
	return errors.New("Could not initialize leveldb")
}
