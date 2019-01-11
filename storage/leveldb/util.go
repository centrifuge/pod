package leveldb

import (
	"fmt"

	"github.com/centrifuge/go-centrifuge/utils"
)

const testStoragePath = "/tmp/centrifuge_data.leveldb_TESTING"

// GetRandomTestStoragePath generates a random path for DB storage
func GetRandomTestStoragePath() string {
	return fmt.Sprintf("%s_%x", testStoragePath, utils.RandomByte32())
}
