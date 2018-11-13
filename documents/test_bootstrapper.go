// +build integration unit

package documents

import (
	"errors"

	"github.com/syndtr/goleveldb/leveldb"
	"github.com/centrifuge/go-centrifuge/storage"
)

// initialized ONLY for tests
var testLevelDB Repository

func (b *Bootstrapper) TestBootstrap(context map[string]interface{}) error {
	if _, ok := context[storage.BootstrappedLevelDb]; !ok {
		return errors.New("initializing LevelDB repository failed")
	}
	testLevelDB = LevelDBRepository{LevelDB: context[storage.BootstrappedLevelDb].(*leveldb.DB)}
	return b.Bootstrap(context)
}

func (*Bootstrapper) TestTearDown() error {
	return nil
}
