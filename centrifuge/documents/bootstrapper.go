package documents

import (
	"errors"

	"github.com/centrifuge/go-centrifuge/centrifuge/bootstrap"
	"github.com/syndtr/goleveldb/leveldb"
)

// initialized ONLY for tests
var testLevelDB Repository

type Bootstrapper struct{}

// Bootstrap sets the required storage and registers
func (*Bootstrapper) Bootstrap(context map[string]interface{}) error {
	if _, ok := context[bootstrap.BootstrappedLevelDb]; !ok {
		return errors.New("initializing LevelDB repository failed")
	}

	testLevelDB = LevelDBRepository{LevelDB: context[bootstrap.BootstrappedLevelDb].(*leveldb.DB)}
	return nil
}

func (b *Bootstrapper) TestBootstrap(context map[string]interface{}) error {
	return b.Bootstrap(context)
}

func (*Bootstrapper) TestTearDown() error {
	return nil
}
