package storage

import (
	"sync"

	"github.com/golang/protobuf/proto"
	logging "github.com/ipfs/go-log"
	"github.com/syndtr/goleveldb/leveldb"
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
			log.Fatal(err)
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

// Getter interface can be implemented by any repository that handles document retrieval
type Getter interface {
	// GetKey will prepare the the identifier key from ID
	GetKey(id []byte) (key []byte)

	// GetByID finds the doc with identifier and marshalls it into message
	GetByID(id []byte, msg proto.Message) error
}

// Checker interface can be implemented by any repository that handles document retrieval
type Checker interface {
	// Exists checks for document existence
	// True if exists else false
	Exists(id []byte) bool
}

// Creator interface can be implemented by any repository that handles document storage
type Creator interface {
	// Create stores the initial document
	// If document exist, it errors out
	Create(id []byte, msg proto.Message) error
}

// Updater interface can be implemented by any repository that handles document storage
type Updater interface {
	// Update updates the already stored document
	// errors out when document is missing
	Update(id []byte, msg proto.Message) error
}

// Repository interface for easy combination
type Repository interface {
	Checker
	Getter
	Creator
	Updater
}
