package storage

import (
	logging "github.com/ipfs/go-log"
	"github.com/syndtr/goleveldb/leveldb"
	"sync"
)

var log = logging.Logger("storage")

var once sync.Once
var instance *leveldb.DB
var dbPath string

// GetStorage is a singleton implementation returning the default database as configured
func NewLeveldbStorage(path string) *leveldb.DB {
	if dbPath != "" {
		log.Fatalf("Can't open new DB, db already open")
	}
	dbPath = path
	once.Do(func() {
		i, err := leveldb.OpenFile(dbPath, nil)
		instance = i
		if err != nil {
			log.Fatalf(err)
		}
	})
	return instance
}

func GetLeveldbStorage() *leveldb.DB {
	return instance
}

func CloseLeveldbStorage() {
	if instance != nil {
		instance.Close()
	}
}
