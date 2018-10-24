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
	assert.NotNil(t, levelDBInstance)
	assert.Equal(t, db, levelDBInstance)

	gdb := GetLevelDBStorage()
	assert.NotNil(t, gdb)
	assert.Equal(t, db, gdb)

	// fail
	db, err = NewLevelDBStorage(path)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "db already open")
	assert.Nil(t, db)

	CloseLevelDBStorage()
	assert.Nil(t, levelDBInstance)
}
