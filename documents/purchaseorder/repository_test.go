// +build unit

package purchaseorder

import (
	"testing"

	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/storage"
	"github.com/stretchr/testify/assert"
)

func TestRepository_getRepository(t *testing.T) {
	r := testRepo()
	assert.NotNil(t, r)
	assert.Equal(t, "purchaseorder", r.(*repository).KeyPrefix)
}

var testRepoGlobal documents.LegacyRepository

func testRepo() documents.LegacyRepository {
	if testRepoGlobal == nil {
		ldb, err := storage.NewLevelDBStorage(storage.GetRandomTestStoragePath())
		if err != nil {
			panic(err)
		}
		testRepoGlobal = getRepository(ldb)
	}
	return testRepoGlobal
}
