// +build integration unit

package documents

import (
	"github.com/centrifuge/go-centrifuge/errors"

	"github.com/centrifuge/go-centrifuge/storage"
)

// initialized ONLY for tests
var testLevelDB Repository

func (b Bootstrapper) TestBootstrap(context map[string]interface{}) error {
	if _, ok := context[storage.BootstrappedDB]; !ok {
		return errors.New("initializing LevelDB repository failed")
	}
	testLevelDB = NewDBRepository(context[storage.BootstrappedDB].(storage.Repository))
	return b.Bootstrap(context)
}

func (Bootstrapper) TestTearDown() error {
	return nil
}
