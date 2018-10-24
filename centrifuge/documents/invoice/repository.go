package invoice

import (
	"github.com/centrifuge/go-centrifuge/centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/centrifuge/storage"
)

// repository is the invoice repository
type repository struct {
	documents.LevelDBRepository
}

// getRepository returns the implemented documents.legacyRepo for invoices
func getRepository() documents.Repository {
	return &repository{
		documents.LevelDBRepository{
			KeyPrefix: "invoice",
			LevelDB:   storage.GetLevelDBStorage(),
		},
	}
}
