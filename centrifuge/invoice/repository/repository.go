package invoicerepository

import (
	"fmt"
	"log"
	"sync"

	"github.com/CentrifugeInc/centrifuge-protobufs/gen/go/invoice"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/code"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/errors"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/invoice"
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

// once to guard from creating multiple instances
var once sync.Once

// InitLevelDBRepository initialises new repository if not exists
func InitLevelDBRepository(db *leveldb.DB) {
	once.Do(func() {
		levelDBRepo = &levelDBRepository{
			storage.DefaultLevelDB{
				KeyPrefix:    "invoice",
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

// validate typecasts and validates the coredocument
func validate(doc proto.Message) error {
	invoiceDoc, ok := doc.(*invoicepb.InvoiceDocument)
	if !ok {
		return errors.New(code.DocumentInvalid, fmt.Sprintf("invalid document of type: %T", doc))
	}

	if valid, msg, errs := invoice.Validate(invoiceDoc); !valid {
		return errors.NewWithErrors(code.DocumentInvalid, msg, errs)
	}

	return nil
}
