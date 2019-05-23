package files

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/stretchr/testify/assert"
	"github.com/syndtr/goleveldb/leveldb"
)

func TestInitial00(t *testing.T) {
	prefix := fmt.Sprintf("/tmp/datadir_%x", utils.RandomByte32())
	targetDir := fmt.Sprintf("%s.leveldb", prefix)

	// Cleanup after test
	defer cleanupDBFiles(prefix)

	db, err := leveldb.OpenFile(targetDir, nil)
	assert.NoError(t, err)

	assert.NoError(t, Initial00(db))
}

func cleanupDBFiles(prefix string) {
	files, err := filepath.Glob(prefix + "*")
	if err != nil {
		panic(err)
	}
	for _, f := range files {
		if err := os.RemoveAll(f); err != nil {
			panic(err)
		}
	}
}
