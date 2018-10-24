package documents

import (
	"fmt"

	"github.com/syndtr/goleveldb/leveldb"
)

// LevelDBRepository is implements repository
type LevelDBRepository struct {
	KeyPrefix string
	LevelDB   *leveldb.DB
}

// Exists returns if the document exists in the repository
func (repo LevelDBRepository) Exists(id []byte) bool {
	_, err := repo.LevelDB.Get(repo.GetKey(id), nil)
	if err != nil {
		return false
	}

	return true
}

// GetKey prepends the id with prefix and returns the result
func (repo LevelDBRepository) GetKey(id []byte) []byte {
	return append([]byte(repo.KeyPrefix), id...)
}

// LoadByID finds the document by id and marshals into message
func (repo LevelDBRepository) LoadByID(id []byte, model Model) error {
	if model == nil {
		return fmt.Errorf("nil document provided")
	}

	data, err := repo.LevelDB.Get(repo.GetKey(id), nil)
	if err != nil {
		return err
	}

	err = model.FromJSON(data)
	if err != nil {
		return err
	}

	return nil
}

// Create creates the document if not exists
// errors out if document exists
func (repo LevelDBRepository) Create(id []byte, model Model) error {
	if model == nil {
		return fmt.Errorf("nil model provided")
	}

	if repo.Exists(id) {
		return fmt.Errorf("document already exists")
	}

	data, err := model.JSON()
	if err != nil {
		return err
	}

	return repo.LevelDB.Put(repo.GetKey(id), data, nil)
}

// Update updates the doc with ID if exists
// errors out if the document
func (repo LevelDBRepository) Update(id []byte, model Model) error {
	if model == nil {
		return fmt.Errorf("nil document provided")
	}

	if !repo.Exists(id) {
		return fmt.Errorf("document doesn't exists")
	}

	data, err := model.JSON()
	if err != nil {
		return err
	}

	return repo.LevelDB.Put(repo.GetKey(id), data, nil)
}
