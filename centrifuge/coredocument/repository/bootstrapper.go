package coredocumentrepository

import (
	"errors"

	"github.com/CentrifugeInc/go-centrifuge/centrifuge/bootstrapper"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/storage"
)

type Bootstrapper struct {
}

func (*Bootstrapper) Bootstrap(context map[string]interface{}) error {
	if _, ok := context[bootstrapper.BootstrappedLevelDb]; ok {
		NewLevelDBRepository(&LevelDBRepository{LevelDB: storage.GetLevelDBStorage()})
		return nil
	}
	return errors.New("could not initialize core document repository")
}
