package migration

import (
	"encoding/json"
	"time"

	"github.com/go-errors/errors"
	"github.com/syndtr/goleveldb/leveldb"
)

const dbPrefix = "migration_"

// Repository holds DB info
type Repository struct {
	db     *leveldb.DB
	dbPath string
}

// Item holds migration item info
type Item struct {
	ID       string        `json:"id"`
	Hash     string        `json:"hash"`
	DateRun  time.Time     `json:"date_run"`
	Duration time.Duration `json:"duration,string"`
}

// NewMigrationRepository takes a path and creates a DB repository
func NewMigrationRepository(path string) (*Repository, error) {
	i, err := leveldb.OpenFile(path, nil)
	if err != nil {
		return nil, err
	}
	return &Repository{i, path}, nil
}

func getKeyFromID(id string) []byte {
	return []byte(dbPrefix + id)
}

// Exists checks that migrationID has been ran
func (repo *Repository) Exists(id string) bool {
	key := getKeyFromID(id)
	res, err := repo.db.Has(key, nil)
	if err != nil {
		return false
	}
	return res
}

// GetMigrationByID returns migration ID if it exists
func (repo *Repository) GetMigrationByID(id string) (*Item, error) {
	v := new(Item)
	key := getKeyFromID(id)
	data, err := repo.db.Get(key, nil)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(data, v)
	if err != nil {
		return nil, err
	}
	return v, nil
}

// CreateMigration stores a migration item in DB
func (repo *Repository) CreateMigration(migrationItem *Item) error {
	if migrationItem == nil {
		return errors.New("nil migration item provided")
	}
	if repo.Exists(migrationItem.ID) {
		return errors.New("migration ID already exists")
	}
	key := getKeyFromID(migrationItem.ID)
	data, err := json.Marshal(migrationItem)
	if err != nil {
		return err
	}
	return repo.db.Put(key, data, nil)
}

// Open opens a DB, requires it to be closed before or it will error out
func (repo *Repository) Open() (err error) {
	repo.db, err = leveldb.OpenFile(repo.dbPath, nil)
	return err
}

// Close closes a DB
func (repo *Repository) Close() (err error) {
	return repo.db.Close()
}
