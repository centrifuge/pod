package context

import (
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/storage"
	"github.com/spf13/viper"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/coredocument/coredocument_repository"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/invoice/invoice_repository"
)

var LevelDB *leveldb.DB

func Bootstrap() {
	path := viper.GetString("storage.Path")
	if path == "" {
		path = "/tmp/centrifuge_data.leveldb_TESTING"
	}
	LevelDB = storage.NewLeveldbStorage(path)

	coredocument_repository.NewLevelDBCoreDocumentRepository(&coredocument_repository.LevelDBCoreDocumentRepository{LevelDB})
	invoice_repository.NewLevelDBInvoiceRepository(&invoice_repository.LevelDBInvoiceRepository{LevelDB})
}
