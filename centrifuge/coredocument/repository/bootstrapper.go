package coredocumentrepository

import (
	"errors"

	"github.com/CentrifugeInc/go-centrifuge/centrifuge/bootstrapper"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/storage"
)

// Bootstrapper implements bootstrapper.Bootstrapper
type Bootstrapper struct{}

// Bootstrap initiates the coredocument repository
func (*Bootstrapper) Bootstrap(context map[string]interface{}) error {
	if _, ok := context[bootstrapper.BootstrappedLevelDb]; ok {
		InitLevelDBRepository(storage.GetLevelDBStorage())
		return nil
	}
	return errors.New("could not initialize core document repository")
}
