package migration

import (
	"encoding/json"
	"github.com/go-errors/errors"
	"time"

	"github.com/syndtr/goleveldb/leveldb"
)

const dbPrefix = "migration_"

type migrationRepo struct {
	db *leveldb.DB
	dbPath string
}

type migrationItem struct {
	ID string              `json:"id"`
	Hash string        		 `json:"hash"`
	DateRun time.Time  		 `json:"date_run,string"`
	Duration time.Duration `json:"duration,string"`
}

func NewMigrationRepository(path string) (*migrationRepo, error) {
	i, err := leveldb.OpenFile(path, nil)
	if err != nil {
		return nil, err
	}
	return &migrationRepo{i, path}, nil
}

func getKeyFromID(id string) []byte {
	return []byte(dbPrefix+id)
}

func (repo *migrationRepo) Exists(id string) bool {
	key := getKeyFromID(id)
	res, err := repo.db.Has(key, nil)
	if err != nil {
		return false
	}
	return res
}

func (repo *migrationRepo) GetMigrationByID(id string) (*migrationItem, error) {
	v := new(migrationItem)
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

func (repo *migrationRepo) CreateMigration(migrationItem *migrationItem) error {
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

func (repo *migrationRepo) DeleteMigration(id string) error {
	if !repo.Exists(id) {
		return errors.New("migration ID does not exist")
	}
	key := getKeyFromID(id)
	return repo.db.Delete(key, nil)
}

func (repo *migrationRepo) RefreshDB() (err error) {
	repo.db, err = leveldb.OpenFile(repo.dbPath, nil)
	return err
}
