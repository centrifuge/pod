package context

import (
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/config"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/coredocument/repository"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/invoice/repository"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/storage"
)

func Bootstrap() {
	config.Config.InitializeViper()
	path := config.Config.GetStoragePath()
	// TODO: it's a bad idea to just write to a test file if the user accidentally configures an empty string as the DB path
	if path == "" {
		path = "/tmp/centrifuge_data.leveldb_TESTING"
	}
	levelDB := storage.NewLeveldbStorage(path)

	coredocumentrepository.NewLevelDBCoreDocumentRepository(&coredocumentrepository.LevelDBCoreDocumentRepository{levelDB})
	invoicerepository.NewLevelDBInvoiceRepository(&invoicerepository.LevelDBInvoiceRepository{levelDB})
}

func Close() {
	levelDB := storage.NewLeveldbStorage(config.Config.GetStoragePath())
	levelDB.Close()
}
