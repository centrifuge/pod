package storage

import (
	"fmt"

	"github.com/centrifuge/go-centrifuge/utils"
)

const testStoragePath = "/tmp/centrifuge_data.leveldb_TESTING"

func GetRandomTestStoragePath() string {
	return fmt.Sprintf("%s_%x", testStoragePath, utils.RandomByte32())
}
