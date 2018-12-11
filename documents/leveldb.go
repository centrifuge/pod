package documents

import (
	"fmt"

	"github.com/syndtr/goleveldb/leveldb"
)

// LegacyRepository should be implemented by any type that wants to store a document in key-value storage.
// Deprecated: Use the single collection DB -> `Repository`
type LegacyRepository interface {
	// GetKey will prepare the the identifier key from ID
	// Deprecated: Use the single collection DB functions -> `Repository`
	GetKey(id []byte) (key []byte)

	// GetByID finds the doc with identifier and marshals it into message
	// Deprecated: Use the single collection DB functions -> `Repository`
	LoadByID(id []byte, model Model) error

	// Exists checks for document existence
	// True if exists else false
	// Deprecated: Use the single collection DB functions -> `Repository`
	Exists(id []byte) bool

	// Create stores the initial document
	// If document exist, it errors out
	// Deprecated: Use the single collection DB functions -> `Repository`
	Create(id []byte, model Model) error

	// Update updates the already stored document
	// errors out when document is missing
	// Deprecated: Use the single collection DB functions -> `Repository`
	Update(id []byte, model Model) error
}

// LevelDBRepository is implements repository
// Deprecated: in favour of `levelDBRepo`
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
