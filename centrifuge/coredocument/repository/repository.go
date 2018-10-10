package coredocumentrepository

import (
	"fmt"
	"log"
	"sync"

	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/go-centrifuge/centrifuge/coredocument"
	"github.com/centrifuge/go-centrifuge/centrifuge/storage"
	"github.com/golang/protobuf/proto"
	"github.com/syndtr/goleveldb/leveldb"
)

// levelDBRepository implements storage.LegacyRepository
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
				KeyPrefix:    "coredoc",
				LevelDB:      db,
				ValidateFunc: validate,
			},
		}
	})
}

// GetRepository returns a repository implementation
// Must be called only after repository initialisation
func GetRepository() storage.LegacyRepository {
	if levelDBRepo == nil {
		log.Fatal("CoreDocument repository not initialised")
	}

	return levelDBRepo
}

// validate typecasts and validates the coredocument
func validate(doc proto.Message) error {
	coreDoc, ok := doc.(*coredocumentpb.CoreDocument)
	if !ok {
		return fmt.Errorf("invalid document type: %T", doc)
	}

	if err := coredocument.Validate(coreDoc); err != nil {
		return err
	}

	return nil
}
