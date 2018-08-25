package coredocumentrepository

import (
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/bootstrapper"
	"github.com/syndtr/goleveldb/leveldb"
	"errors"
)

type Bootstrapper struct {
}

func (*Bootstrapper) Bootstrap(context map[string]interface{}) error {
	if levelDb, ok := context[bootstrapper.BOOTSTRAPPED_LEVELDB]; ok {
		if typedLevelDb, castok := levelDb.(*leveldb.DB); castok {
			NewLevelDBRepository(&LevelDBRepository{LevelDB: typedLevelDb})
			return nil
		}
	}
	return errors.New("could not initialize core document repository")
}
