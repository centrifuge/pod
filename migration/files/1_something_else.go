package files

import (
	"github.com/go-errors/errors"
	"github.com/syndtr/goleveldb/leveldb"
)

// RunMigration1 something else
func RunMigration1(db *leveldb.DB) error {
	err := db.Put([]byte("damn"), []byte("alot"), nil)
	if err != nil {
		return err
	}
	log.Errorf("Migration 1 Run failed")
	return errors.New("Something failed")
}
