//go:build unit

package migrationfiles

import (
	"fmt"
	"path"
	"testing"

	migrationutils "github.com/centrifuge/go-centrifuge/migration/utils"
	testingcommons "github.com/centrifuge/go-centrifuge/testingutils/common"
	"github.com/stretchr/testify/assert"
	"github.com/syndtr/goleveldb/leveldb"
)

const (
	migrationFilesTestDirPattern = "migration-files-test-*"
)

func TestInitial00(t *testing.T) {
	testDir, err := testingcommons.GetRandomTestStoragePath(migrationFilesTestDirPattern)
	assert.NoError(t, err)

	defer migrationutils.CleanupDBFiles(testDir)

	dbFileName := fmt.Sprintf("%x.leveldb", migrationutils.RandomByte32())

	targetDir := path.Join(testDir, dbFileName)

	db, err := leveldb.OpenFile(targetDir, nil)
	assert.NoError(t, err)

	assert.NoError(t, Initial00(db))
}
