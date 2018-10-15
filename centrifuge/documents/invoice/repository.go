package invoice

import (
	"fmt"
	"sync"

	"github.com/centrifuge/centrifuge-protobufs/gen/go/invoice"
	"github.com/centrifuge/go-centrifuge/centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/centrifuge/storage"
	"github.com/golang/protobuf/proto"
	"github.com/syndtr/goleveldb/leveldb"
)

// legacyLevelDBRepo implements storage.LegacyRepository
// This is a legacy repository
type legacyLevelDBRepo struct {
	storage.DefaultLevelDB
}

// legacyRepo is singleton instance
var legacyRepo *legacyLevelDBRepo

// once to guard from creating multiple instances
var once sync.Once

// InitLegacyRepository initialises new repository if not exists
func InitLegacyRepository(db *leveldb.DB) {
	once.Do(func() {
		legacyRepo = &legacyLevelDBRepo{
			storage.DefaultLevelDB{
				KeyPrefix:    "invoice",
				LevelDB:      db,
				ValidateFunc: validate,
			},
		}
	})
}

// GetLegacyRepository returns a repository implementation
// Must be called only after repository initialisation
func GetLegacyRepository() storage.LegacyRepository {
	if legacyRepo == nil {
		log.Fatal("Invoice repository not initialised")
	}

	return legacyRepo
}

// validate typecasts and validates the coredocument
func validate(doc proto.Message) error {
	invoiceDoc, ok := doc.(*invoicepb.InvoiceDocument)
	if !ok {
		return fmt.Errorf("invalid document of type: %T", doc)
	}

	if err := Validate(invoiceDoc); err != nil {
		return err
	}

	return nil
}

// repository is the invoice repository
type repository struct {
	documents.LevelDBRepository
}

var repo *repository

// GetRepository returns the implemented documents.legacyRepo for invoices
func GetRepository() documents.Repository {
	if repo == nil {
		repo = &repository{
			documents.LevelDBRepository{
				KeyPrefix: "invoice",
				LevelDB:   storage.GetLevelDBStorage(),
			},
		}
	}

	return repo
}
