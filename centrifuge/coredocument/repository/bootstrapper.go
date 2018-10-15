package coredocumentrepository

import (
	"errors"

	"github.com/centrifuge/go-centrifuge/centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/centrifuge/storage"
)

// Bootstrapper implements bootstrapper.Bootstrapper
type Bootstrapper struct{}

// Bootstrap initiates the coredocument repository
func (*Bootstrapper) Bootstrap(context map[string]interface{}) error {
	if _, ok := context[bootstrap.BootstrappedLevelDb]; ok {
		InitLevelDBRepository(storage.GetLevelDBStorage())
		return nil
	}
	return errors.New("could not initialize core document repository")
}

func (b *Bootstrapper) TestBootstrap(context map[string]interface{}) error {
	return b.Bootstrap(context)
}

func (*Bootstrapper) TestTearDown() error {
	return nil
}
