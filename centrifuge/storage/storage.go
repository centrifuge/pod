package storage

import (
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/spf13/viper"
	"sync"
)

type configStruct struct {
	path string
}

type LeveldbDataStore struct {
	leveldb *leveldb.DB
	config configStruct
}

func (db *LeveldbDataStore) Open() (err error) {
	newLeveldb, err := leveldb.OpenFile(db.config.path, nil)
	db.leveldb = newLeveldb;
	return err
}

func (db *LeveldbDataStore) Close() {
	db.leveldb.Close()
}

func (db *LeveldbDataStore) Put(key []byte, doc []byte) (err error) {
	err = db.leveldb.Put(key, doc, nil)
	return
}

func (db *LeveldbDataStore) Get(key []byte) (doc []byte, err error) {
	doc, err = db.leveldb.Get(key, nil)
	return
}

var once sync.Once
var instance DataStore

func GetStorage() *DataStore {
	once.Do(func() {
		instance = &LeveldbDataStore{config: configStruct{path: viper.GetString("storage.path")}}
		err := instance.Open()
		if err != nil {
			panic(err)
		}

	})
	return &instance
}