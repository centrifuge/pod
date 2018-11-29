package invoice

import (
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/syndtr/goleveldb/leveldb"
)

// repository is the invoice repository
type repository struct {
	documents.LevelDBRepository
}

// getRepository returns the implemented documents.legacyRepo for invoices
func getRepository(ldb *leveldb.DB) documents.Repository {
	return &repository{
		documents.LevelDBRepository{
			KeyPrefix: "invoice",
			LevelDB:   ldb,
		},
	}
}
