package migrationfiles

import (
	"fmt"
	"testing"

	"github.com/centrifuge/go-centrifuge/migration/utils"
	"github.com/stretchr/testify/assert"
	"github.com/syndtr/goleveldb/leveldb"
)

func TestInitial00(t *testing.T) {
	prefix := fmt.Sprintf("/tmp/datadir_%x", migrationutils.RandomByte32())
	targetDir := fmt.Sprintf("%s.leveldb", prefix)

	// Cleanup after test
	defer migrationutils.CleanupDBFiles(prefix)

	db, err := leveldb.OpenFile(targetDir, nil)
	assert.NoError(t, err)

	assert.NoError(t, Initial00(db))
}
