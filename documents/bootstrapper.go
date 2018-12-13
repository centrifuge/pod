package documents

import (
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/storage"
	"github.com/syndtr/goleveldb/leveldb"
)

// BootstrappedRegistry is the key to ServiceRegistry in Bootstrap context
const BootstrappedRegistry = "BootstrappedRegistry"

// BootstrappedDocumentRepository is the key to the database repository of documents
const BootstrappedDocumentRepository = "BootstrappedDocumentRepository"

// Bootstrapper implements bootstrap.Bootstrapper.
type Bootstrapper struct{}

// Bootstrap sets the required storage and registers
func (Bootstrapper) Bootstrap(ctx map[string]interface{}) error {
	ctx[BootstrappedRegistry] = NewServiceRegistry()
	ldb, ok := ctx[storage.BootstrappedLevelDB].(*leveldb.DB)
	if !ok {
		return errors.Error(ErrDocumentBootstrap)
	}
	repo := NewLevelDBRepository(ldb)
	ctx[BootstrappedDocumentRepository] = repo
	return nil
}
