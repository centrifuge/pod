package purchaseorder

import (
	"fmt"
	"sync"

	"github.com/centrifuge/centrifuge-protobufs/gen/go/purchaseorder"
	"github.com/centrifuge/go-centrifuge/centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/centrifuge/storage"
	"github.com/golang/protobuf/proto"
	"github.com/syndtr/goleveldb/leveldb"
)

// levelDBRepository implements storage.LegacyRepository
// Deprecated
type levelDBRepository struct {
	storage.DefaultLevelDB
}

// levelDBRepo is singleton instance
// Deprecated
var levelDBRepo *levelDBRepository

// legacyOnce to guard from creating multiple instances
var legacyOnce sync.Once

// InitLegacyLevelDBRepository initialises new repository if not exists
// Deprecated
func InitLegacyLevelDBRepository(db *leveldb.DB) {
	legacyOnce.Do(func() {
		levelDBRepo = &levelDBRepository{
			storage.DefaultLevelDB{
				KeyPrefix:    "purchaseorder",
				LevelDB:      db,
				ValidateFunc: validate,
			},
		}
	})
}

// GetLegacyRepository returns a repository implementation
// Must be called only after repository initialisation
// Deprecated
func GetLegacyRepository() storage.LegacyRepository {
	if levelDBRepo == nil {
		log.Fatal("Invoice repository not initialised")
	}

	return levelDBRepo
}

func validate(doc proto.Message) error {
	poDoc, ok := doc.(*purchaseorderpb.PurchaseOrderDocument)
	if !ok {
		return fmt.Errorf("invalid document of type: %T", doc)
	}

	if err := Validate(poDoc); err != nil {
		return err
	}

	return nil
}

// repository is the invoice repository
type repository struct {
	documents.LevelDBRepository
}

// repo is a singleton instance of repository
var repo *repository

// once guards the singleton initialisation
var once sync.Once

// getRepository returns the implemented documents.legacyRepo for invoices
func getRepository() documents.Repository {
	once.Do(func() {
		repo = &repository{
			documents.LevelDBRepository{
				KeyPrefix: "purchaseorder",
				LevelDB:   storage.GetLevelDBStorage(),
			},
		}
	})

	return repo
}
