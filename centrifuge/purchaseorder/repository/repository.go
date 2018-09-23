package purchaseorderrepository

import (
	"log"
	"sync"

	"fmt"

	"github.com/centrifuge/centrifuge-protobufs/gen/go/purchaseorder"
	"github.com/centrifuge/go-centrifuge/centrifuge/centerrors"
	"github.com/centrifuge/go-centrifuge/centrifuge/code"
	"github.com/centrifuge/go-centrifuge/centrifuge/purchaseorder"
	"github.com/centrifuge/go-centrifuge/centrifuge/storage"
	"github.com/golang/protobuf/proto"
	"github.com/syndtr/goleveldb/leveldb"
)

// levelDBRepository implements storage.Repository
type levelDBRepository struct {
	storage.DefaultLevelDB
}

// levelDBRepo is singleton instance
var levelDBRepo *levelDBRepository

// once to guard from creating multiple instances
var once sync.Once

// InitLevelDBRepository initialises new repository if not exists
func InitLevelDBRepository(db *leveldb.DB) {
	once.Do(func() {
		levelDBRepo = &levelDBRepository{
			storage.DefaultLevelDB{
				KeyPrefix:    "purchaseorder",
				LevelDB:      db,
				ValidateFunc: validate,
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

func validate(doc proto.Message) error {
	poDoc, ok := doc.(*purchaseorderpb.PurchaseOrderDocument)
	if !ok {
		return centerrors.New(code.DocumentInvalid, fmt.Sprintf("invalid document of type: %T", doc))
	}

	if valid, msg, errs := purchaseorder.Validate(poDoc); !valid {
		return centerrors.NewWithErrors(code.DocumentInvalid, msg, errs)
	}

	return nil
}
