package storage

import (
	"github.com/syndtr/goleveldb/leveldb"
	"sync"
)

var once sync.Once

// GetStorage is a singleton implementation returning the default database as configured
func NewLeveldbStorage(path string) *leveldb.DB {
	var instance *leveldb.DB
	once.Do(func() {
		i, err := leveldb.OpenFile(path, nil)
		instance = i
		if err != nil {
			panic(err)
		}
	})
	return instance
}
