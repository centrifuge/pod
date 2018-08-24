package storage

import (
	"sync"

	"fmt"

	"github.com/golang/protobuf/proto"
	logging "github.com/ipfs/go-log"
	"github.com/syndtr/goleveldb/leveldb"
)

var log = logging.Logger("storage")

// levelDBInstance is levelDB instance
var levelDBInstance *leveldb.DB

// once to guard the levelDB instance
var once sync.Once

// GetStorage is a singleton implementation returning the default database as configured
func NewLevelDBStorage(path string) *leveldb.DB {
	if levelDBInstance != nil {
		log.Fatalf("Can't open new DB, db already open")
	}

	once.Do(func() {
		i, err := leveldb.OpenFile(path, nil)
		if err != nil {
			log.Fatal(err)
		}

		levelDBInstance = i
	})
	return levelDBInstance
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
	if levelDBInstance != nil {
		levelDBInstance.Close()
	}
}

// DefaultLevelDB implements the repository
type DefaultLevelDB struct {
	KeyPrefix    string
	LevelDB      *leveldb.DB
	ValidateFunc func([]byte, proto.Message) error
}

// Exists returns if the document exists in the repository
func (repo *DefaultLevelDB) Exists(id []byte) bool {
	_, err := repo.LevelDB.Get(repo.GetKey(id), nil)
	if err != nil {
		return false
	}

	return true
}

// GetKey prepends the id with prefix and returns the result
func (repo *DefaultLevelDB) GetKey(id []byte) []byte {
	return append([]byte(repo.KeyPrefix), id...)
}

// GetByID finds the document by id and marshalls into message
func (repo *DefaultLevelDB) GetByID(id []byte, msg proto.Message) error {
	if msg == nil {
		return fmt.Errorf("nil document provided")
	}

	data, err := repo.LevelDB.Get(repo.GetKey(id), nil)
	if err != nil {
		return err
	}

	err = proto.Unmarshal(data, msg)
	if err != nil {
		return err
	}

	return nil
}

// Create creates the document if not exists
// errors out if document exist
func (repo *DefaultLevelDB) Create(id []byte, msg proto.Message) error {
	if msg == nil {
		return fmt.Errorf("nil document provided")
	}

	if repo.Exists(id) {
		return fmt.Errorf("document already exists")
	}

	if repo.ValidateFunc != nil {
		err := repo.ValidateFunc(id, msg)
		if err != nil {
			return fmt.Errorf("validation failed: %v", err)
		}
	}

	data, err := proto.Marshal(msg)
	if err != nil {
		return err
	}

	return repo.LevelDB.Put(repo.GetKey(id), data, nil)
}

// Update updates the doc with ID if exists
func (repo *DefaultLevelDB) Update(id []byte, msg proto.Message) error {
	if msg == nil {
		return fmt.Errorf("nil document provided")
	}

	if !repo.Exists(id) {
		return fmt.Errorf("document doesn't exists")
	}

	if repo.ValidateFunc != nil {
		err := repo.ValidateFunc(id, msg)
		if err != nil {
			return fmt.Errorf("validation failed: %v", err)
		}
	}

	data, err := proto.Marshal(msg)
	if err != nil {
		return err
	}

	return repo.LevelDB.Put(repo.GetKey(id), data, nil)
}
