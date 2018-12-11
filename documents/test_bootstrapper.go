// +build integration unit

package documents

import (
	"errors"

	"github.com/centrifuge/go-centrifuge/config"

	"github.com/syndtr/goleveldb/leveldb"
)

// initialized ONLY for tests
var testLevelDB LegacyRepository

func (b Bootstrapper) TestBootstrap(context map[string]interface{}) error {
	if _, ok := context[config.BootstrappedLevelDB]; !ok {
		return errors.New("initializing LevelDB repository failed")
	}
	testLevelDB = LevelDBRepository{LevelDB: context[config.BootstrappedLevelDB].(*leveldb.DB)}
	return b.Bootstrap(context)
}

func (Bootstrapper) TestTearDown() error {
	return nil
}
