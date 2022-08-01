package leveldb

import (
	"os"
)

const (
	storageDirPattern = "go-centrifuge-test-*"
)

// GetRandomTestStoragePath generates a random path for DB storage
func GetRandomTestStoragePath() (string, error) {
	return os.MkdirTemp(os.TempDir(), storageDirPattern)
}
