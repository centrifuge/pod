package context

import (
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/coredocument/repository"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/invoice/repository"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/networks"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/storage"
	"github.com/spf13/viper"
)

func Bootstrap() {
	path := viper.GetString("storage.Path")
	if path == "" {
		path = "/tmp/centrifuge_data.leveldb_TESTING"
	}
	levelDB := storage.NewLeveldbStorage(path)

	coredocumentrepository.NewLevelDBCoreDocumentRepository(&coredocumentrepository.LevelDBCoreDocumentRepository{levelDB})
	invoicerepository.NewLevelDBInvoiceRepository(&invoicerepository.LevelDBInvoiceRepository{levelDB})
	networks.InitViperNetworkConfigurationLoader(networks.NewViperNetworkConfigurationLoader())
}
