// +build unit

package storage

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewLevelDBStorage(t *testing.T) {
	path := getRandomTestStoragePath()
	db, err := NewLevelDBStorage(path)
	assert.Nil(t, err)
	assert.NotNil(t, db)
}
