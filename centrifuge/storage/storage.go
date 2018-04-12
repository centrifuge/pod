package storage

import (
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/spf13/viper"
	"sync"
	"github.com/CentrifugeInc/centrifuge-protobufs/coredocument"
	"github.com/golang/protobuf/proto"
)


type LeveldbDataStore struct {
	leveldb *leveldb.DB
	Path    string
}

// Open the database connection
func (db *LeveldbDataStore) Open() (err error) {
	newLeveldb, err := leveldb.OpenFile(db.Path, nil)
	db.leveldb = newLeveldb;
	return err
}

// Close the database connection when the program terminates
func (db *LeveldbDataStore) Close() {
	db.leveldb.Close()
}

// Put a document in store without checking if it already exists. Expects a byte array key
func (db *LeveldbDataStore) Put(key []byte, doc []byte) (err error) {
	err = db.leveldb.Put(key, doc, nil)
	return
}

// Get a document from storage. Returns an error if it does not exist
func (db *LeveldbDataStore) Get(key []byte) (doc []byte, err error) {
	doc, err = db.leveldb.Get(key, nil)
	return
}

var (
	once sync.Once
	instance DataStore
)

func (db *LeveldbDataStore) GetDocumentKey(id []byte) (key []byte) {
	key = append([]byte("coredoc"), id...)
	return key
}

func (db *LeveldbDataStore) GetDocument(id []byte) (doc *coredocument.CoreDocument, err error) {
	doc_bytes, err := db.Get(db.GetDocumentKey(id))
	if err != nil {
		return nil, err
	}

	doc = &coredocument.CoreDocument{}
	err = proto.Unmarshal(doc_bytes, doc)
	if err != nil {
		return nil, err
	}
	return
}

func (db *LeveldbDataStore) PutDocument(doc *coredocument.CoreDocument) (err error) {
	key := db.GetDocumentKey(doc.DocumentIdentifier)
	data, err := proto.Marshal(doc)

	if err != nil {
		return
	}
	err = db.Put(key, data)
	return
}


// GetStorage is a singleton implementation returning the default database as configured
func GetLeveldbStorage() DataStore {
	once.Do(func() {
		if instance != nil {
			return
		}
		path := viper.GetString("storage.Path")
		if path == "" {
			path = "/tmp/centrifuge_data.leveldb_TESTING"
		}
		instance = &LeveldbDataStore{Path: path}
		err := instance.Open()
		if err != nil {
			panic(err)
		}
	})
	return instance
}

// SetStorage can be used to overwrite the default database with something else for testing purposes.
func SetStorage (store DataStore) {
	if instance != nil {
		panic("Can't set storage, storage already instantiated")
	}
	instance = store
}

