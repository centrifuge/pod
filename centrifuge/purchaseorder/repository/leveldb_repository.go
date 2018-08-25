package purchaseorderrepository

import (
	"fmt"
	"log"
	"sync"

	"github.com/CentrifugeInc/centrifuge-protobufs/gen/go/purchaseorder"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/errors"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/storage"
	"github.com/golang/protobuf/proto"
	"github.com/syndtr/goleveldb/leveldb"
)

// levelDBRepository implements storage.Repository
type levelDBRepository struct {
	storage.DefaultLevelDB
}

// levelDBRepo is singleton instance
var levelDBRepo *levelDBRepository

// once to guard from multiple instances
var once sync.Once

// InitLevelDBRepository initialises new repository if not exists
func InitLevelDBRepository(db *leveldb.DB) {
	once.Do(func() {
		levelDBRepo = &levelDBRepository{
			storage.DefaultLevelDB{
				KeyPrefix:    "purchaseorder",
				LevelDB:      db,
				ValidateFunc: checkIfCoreDocumentFilledCorrectly,
			},
		}
	})
}

// GetRepository returns a repository implementation
// Must be called only after repository initialisation
func GetRepository() storage.Repository {
	if levelDBRepo == nil {
		log.Fatal("Invoice repository not initialised")
	}

	return levelDBRepo
}

// checkIfCoreDocumentFilledCorrectly checks if the core document details are filled
func checkIfCoreDocumentFilledCorrectly(_ []byte, msg proto.Message) error {
	doc, ok := msg.(*purchaseorderpb.PurchaseOrderDocument)
	if !ok {
		return fmt.Errorf("unrecognized type: %T", msg)
	}

	if doc.CoreDocument == nil {
		return errors.NilError(doc.CoreDocument)
	}

	if doc.CoreDocument.DocumentIdentifier == nil {
		return errors.NilError(doc.CoreDocument.DocumentIdentifier)
	}

	return nil
}
