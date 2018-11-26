package storage

import (
	"fmt"
	"sync"

	logging "github.com/ipfs/go-log"
	"github.com/syndtr/goleveldb/leveldb"
)

var log = logging.Logger("storage")

// levelDBInstance is levelDB instance
var levelDBInstance *leveldb.DB

// lock to guard the levelDB instance
var lock sync.Mutex

// NewLevelDBStorage is a singleton implementation returning the default database as configured.
func NewLevelDBStorage(path string) (*leveldb.DB, error) {
	if levelDBInstance != nil {
		return nil, fmt.Errorf("db already open")
	}

	lock.Lock()
	defer lock.Unlock()
	i, err := leveldb.OpenFile(path, nil)
	if err != nil {
		log.Fatal(err)
	}

	levelDBInstance = i
	return levelDBInstance, nil
}

// GetLevelDBStorage returns levelDB instance if initialised
// panics if not initialised
func GetLevelDBStorage() *leveldb.DB {
	if levelDBInstance == nil {
		log.Fatalf("LevelDB not initialised")
	}

	return levelDBInstance
}

// CloseLevelDBStorage closes any open instance of levelDB
func CloseLevelDBStorage() {
	if levelDBInstance == nil {
		return
	}

	lock.Lock()
	defer lock.Unlock()
	err := levelDBInstance.Close()
	if err != nil {
		log.Infof("failed to close the level DB: %v", err)
	}

	levelDBInstance = nil
}
