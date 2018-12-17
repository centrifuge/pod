package storage

import (
	logging "github.com/ipfs/go-log"
	"github.com/syndtr/goleveldb/leveldb"
)

var log = logging.Logger("storage")

// NewLevelDBStorage is a singleton implementation returning the default database as configured.
func NewLevelDBStorage(path string) (*leveldb.DB, error) {
	i, err := leveldb.OpenFile(path, nil)
	if err != nil {
		return nil, err
	}
	return i, nil
}
