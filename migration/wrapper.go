package migration

// Package the default resources into binary data that is embedded in centrifuge
// executable
//
//go:generate go-bindata -pkg migration -prefix "../" -o ./data.go ./files/...

import (
	"crypto/sha256"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"sort"
	"strings"
	"time"

	"github.com/centrifuge/go-centrifuge/migration/files"
	"github.com/ethereum/go-ethereum/common/hexutil"
	logging "github.com/ipfs/go-log"
	"github.com/syndtr/goleveldb/leveldb"
)

var log = logging.Logger("migrate-cmd")

var migrations = map[string]func(*leveldb.DB) error{
	"0InitialMigration": files.RunMigration0,
}

// Runner is the actor that runs the migrations
type Runner struct {
	CalculateHash func(string) (string, error)
}

// NewMigrationRunner creates default runner
func NewMigrationRunner() *Runner {
	return NewMigrationRunnerWithHashFunction(calculateMigrationHash)
}

// NewMigrationRunnerWithHashFunction creates runner with custom hash function
func NewMigrationRunnerWithHashFunction(hashFn func(string) (string, error)) *Runner {
	return &Runner{hashFn}
}

// RunMigrations executes the migrations
func (mr *Runner) RunMigrations(dbPath string) error {
	repo, err := NewMigrationRepository(dbPath)
	if err != nil {
		return err
	}

	var bkpRepo *Repository
	migrationList := make([]string, 0, len(migrations))
	for k := range migrations {
		migrationList = append(migrationList, k)
	}
	sort.Strings(migrationList)

	//For each of them, in order execute
	for _, k := range migrationList {
		start := time.Now()
		// Add hash check
		if repo.Exists(k) {
			log.Infof("Migration %s already run", k)
			continue
		}

		// backup DB
		bkpRepo, err = backupDB(repo, k)
		if err != nil {
			return nil
		}

		// execute migration file
		if err = migrations[k](repo.db); err != nil {
			log.Errorf("Migration %s failed", k)
			err1 := revertDBToBackup(repo, bkpRepo)
			if err1 != nil {
				return err1
			}
			return err
		}

		// on success store migration item in DB
		hash, err := mr.CalculateHash(k)
		if err != nil {
			err1 := revertDBToBackup(repo, bkpRepo)
			if err1 != nil {
				return err1
			}
			return err
		}
		mi := &Item{
			ID:       k,
			DateRun:  time.Now().UTC(),
			Duration: time.Now().Sub(start),
			Hash:     hash,
		}
		if err = repo.CreateMigration(mi); err != nil {
			err1 := revertDBToBackup(repo, bkpRepo)
			if err1 != nil {
				return err1
			}
			return err
		}

		err = bkpRepo.db.Close()
		if err != nil {
			return err
		}

	}

	return repo.db.Close()
}

func calculateMigrationHash(name string) (string, error) {
	data, err := Asset(fmt.Sprintf("migration/files/%s.go", name))
	if err != nil {
		return "", err
	}
	hb, err := sha256Hash(data)
	if err != nil {
		return "", err
	}
	return hexutil.Encode(hb), nil
}

func getBackupName(path, name string) string {
	bkpPath := strings.TrimSuffix(path, ".leveldb")
	return fmt.Sprintf("%s_%s.leveldb", bkpPath, name)
}

func backupDB(srcRepo *Repository, migrationID string) (bkp *Repository, err error) {
	//Closing src to make backup
	err = srcRepo.db.Close()
	if err != nil {
		return nil, err
	}

	dstPath := getBackupName(srcRepo.dbPath, migrationID)
	err = CopyDir(srcRepo.dbPath, dstPath)
	if err != nil {
		return nil, err
	}

	// Refreshing src DB
	err = srcRepo.RefreshDB()
	if err != nil {
		return nil, err
	}

	return NewMigrationRepository(dstPath)
}

func revertDBToBackup(srcDB, bkpDB *Repository) error {
	erri := srcDB.db.Close()
	if erri != nil {
		return erri
	}
	erri = bkpDB.db.Close()
	if erri != nil {
		return erri
	}

	err := os.RemoveAll(srcDB.dbPath)
	if err != nil {
		return err
	}
	return os.Rename(bkpDB.dbPath, srcDB.dbPath)
}

// CopyFile copies a single file from src to dst
// Code taken from https://blog.depado.eu/post/copy-files-and-directories-in-go
func CopyFile(src, dst string) error {
	var err error
	var srcfd *os.File
	var dstfd *os.File
	var srcinfo os.FileInfo

	if srcfd, err = os.Open(src); err != nil {
		return err
	}
	defer srcfd.Close()

	if dstfd, err = os.Create(dst); err != nil {
		return err
	}
	defer dstfd.Close()

	if _, err = io.Copy(dstfd, srcfd); err != nil {
		return err
	}
	if srcinfo, err = os.Stat(src); err != nil {
		return err
	}
	return os.Chmod(dst, srcinfo.Mode())
}

// CopyDir copies a whole directory recursively
func CopyDir(src string, dst string) error {
	var err error
	var fds []os.FileInfo
	var srcinfo os.FileInfo

	if srcinfo, err = os.Stat(src); err != nil {
		return err
	}

	if err = os.MkdirAll(dst, srcinfo.Mode()); err != nil {
		return err
	}

	if fds, err = ioutil.ReadDir(src); err != nil {
		return err
	}
	for _, fd := range fds {
		srcfp := path.Join(src, fd.Name())
		dstfp := path.Join(dst, fd.Name())

		if fd.IsDir() {
			if err = CopyDir(srcfp, dstfp); err != nil {
				fmt.Println(err)
			}
		} else {
			if err = CopyFile(srcfp, dstfp); err != nil {
				fmt.Println(err)
			}
		}
	}
	return nil
}

// sha256Hash wraps inconvenient sha256 hashing ops
func sha256Hash(value []byte) (hash []byte, err error) {
	h := sha256.New()
	_, err = h.Write(value)
	if err != nil {
		return []byte{}, err
	}
	return h.Sum(nil), nil
}
