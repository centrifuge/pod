package purchaseorder

import (
	"github.com/centrifuge/go-centrifuge/centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/centrifuge/storage"
)

// repository is the purchase order repository
type repository struct {
	documents.LevelDBRepository
}

// getRepository returns the implemented documents.legacyRepo for purchase orders
func getRepository() documents.Repository {
	return &repository{
		documents.LevelDBRepository{
			KeyPrefix: "purchaseorder",
			LevelDB:   storage.GetLevelDBStorage(),
		},
	}
}
