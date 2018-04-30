package context

import (
	"sync"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/storage"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/spf13/viper"
)

var (
	once sync.Once
	LevelDB *leveldb.DB
)

func Bootstrap() {
	once.Do(func() {
		path := viper.GetString("storage.Path")
		if path == "" {
			path = "/tmp/centrifuge_data.leveldb_TESTING"
		}
		LevelDB = storage.GetLeveldbStorage(path)
	})
}
