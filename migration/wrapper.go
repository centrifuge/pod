package migration

// Package the default resources into binary data that is embedded in centrifuge
// executable
//
//go:generate go-bindata -pkg migration -prefix "../" -o ./data.go ./files/...

import (
	"crypto/sha256"
	"fmt"
	"github.com/centrifuge/go-centrifuge/migration/files"
	"github.com/ethereum/go-ethereum/common/hexutil"
	logging "github.com/ipfs/go-log"
	"github.com/syndtr/goleveldb/leveldb"
	"io"
	"io/ioutil"
	"os"
	"path"
	"sort"
	"strings"
	"time"
)

var log = logging.Logger("migrate-cmd")

var migrations = map[string]func(*leveldb.DB)error{
	"0_job_key_to_hex": files.RunMigration0,
	"1_something_else": files.RunMigration1,
}

type Runner struct {
	CalculateHash func (string) (string, error)
}

func NewMigrationRunner() *Runner {
	return NewMigrationRunnerWithHashFunction(calculateMigrationHash)
}

func NewMigrationRunnerWithHashFunction(hashFn func(string)(string, error)) *Runner {
	return &Runner{hashFn}
}

func (mr *Runner) RunMigrations(dbPath string) error {
	repo, err := NewMigrationRepository(dbPath)
	if err != nil {
		return err
	}

	var bkp *leveldb.DB
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
		//Closing to make backup
		err = repo.db.Close()
		if err != nil {
			return err
		}

		// backup DB
		bkp, err = backupDB(dbPath, k)
		if err != nil {
			return nil
		}
		// execute migration file
		err = repo.RefreshDB()
		if err != nil {
			return nil
		}
		if err = migrations[k](repo.db); err != nil {
			log.Errorf("Migration %s failed", k)
			erri := repo.db.Close()
			if erri != nil {
				return erri
			}
			erri = bkp.Close()
			if erri != nil {
				return erri
			}
			erri = revertToBackup(getBackupName(dbPath, k), dbPath)
			if erri != nil {
				return erri
			}
			return err
		}

		// else store migration in DB
		hash, err := mr.CalculateHash(k)
		if err != nil {
			repo.db.Close()
			erri := revertToBackup(getBackupName(dbPath, k), dbPath)
			if erri != nil {
				return erri
			}
			return err
		}
		mi := &migrationItem{
			ID:       k,
			DateRun:  time.Now().UTC(),
			Duration: time.Now().Sub(start),
			Hash:     hash,
		}
		if err = repo.CreateMigration(mi); err != nil {
			repo.db.Close()
			erri := revertToBackup(getBackupName(dbPath, k), dbPath)
			if erri != nil {
				return erri
			}
			return err
		}

		err = bkp.Close()
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

func backupDB(srcPath, migrationID string) (bkp *leveldb.DB, err error) {
	dstPath := getBackupName(srcPath, migrationID)
	err = CopyDir(srcPath, dstPath)
	if err != nil {
		return nil, err
	}

	return leveldb.OpenFile(dstPath, nil)
}

func revertToBackup(bkpPath, srcPath string) error {
	err := os.RemoveAll(srcPath)
	if err != nil {
		return err
	}
	return os.Rename(bkpPath, srcPath)
}

// Code taken from https://blog.depado.eu/post/copy-files-and-directories-in-go
// File copies a single file from src to dst
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

// Dir copies a whole directory recursively
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

// Sha256Hash wraps inconvenient sha256 hashing ops
func sha256Hash(value []byte) (hash []byte, err error) {
	h := sha256.New()
	_, err = h.Write(value)
	if err != nil {
		return []byte{}, err
	}
	return h.Sum(nil), nil
}