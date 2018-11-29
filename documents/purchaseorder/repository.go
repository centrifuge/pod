package purchaseorder

import (
	"github.com/centrifuge/go-centrifuge/documents"
		"github.com/syndtr/goleveldb/leveldb"
)

// repository is the purchase order repository
type repository struct {
	documents.LevelDBRepository
}

// getRepository returns the implemented documents.legacyRepo for purchase orders
func getRepository(ldb *leveldb.DB) documents.Repository {
	return &repository{
		documents.LevelDBRepository{
			KeyPrefix: "purchaseorder",
			LevelDB:   ldb,
		},
	}
}
