package purchaseorderrepository

import (
	"errors"

	"github.com/CentrifugeInc/go-centrifuge/centrifuge/bootstrapper"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/storage"
)

type Bootstrapper struct{}

func (*Bootstrapper) Bootstrap(context map[string]interface{}) error {
	if _, ok := context[bootstrapper.BootstrappedLevelDb]; ok {
		InitLevelDBRepository(storage.GetLevelDBStorage())
		return nil
	}
	return errors.New("could not initialize purchase order repository")
}
