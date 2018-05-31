package storage

import (
	"github.com/syndtr/goleveldb/leveldb"
	"sync"
)

var once sync.Once
var instance *leveldb.DB

// GetStorage is a singleton implementation returning the default database as configured
func NewLeveldbStorage(path string) *leveldb.DB {
	// TODO: I don't like how the second invocation of this method completely
	// ignores the path. If at any point in time the db gets initialized with a
	// different path, then bad stuff can happen.

	once.Do(func() {
		i, err := leveldb.OpenFile(path, nil)
		instance = i
		if err != nil {
			panic(err)
		}
	})
	return instance
}
