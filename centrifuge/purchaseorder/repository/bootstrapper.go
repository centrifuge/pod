package purchaseorderrepository

import (
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/bootstrapper"
	"github.com/syndtr/goleveldb/leveldb"
)

type Bootstrapper struct {
}

func (*Bootstrapper) Bootstrap(context map[string]interface{}) error {
	if levelDb, ok := context[bootstrapper.BOOTSTRAPPED_LEVELDB]; ok {
		if typedLevelDb, castok := levelDb.(*leveldb.DB); castok {
			InitLevelDBRepository(typedLevelDb)
			return nil
		}
	}
	return nil
}
