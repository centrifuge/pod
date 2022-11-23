package migration

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"sort"
	"strings"
	"time"

	logging "github.com/ipfs/go-log"
	"github.com/syndtr/goleveldb/leveldb"
)

var log = logging.Logger("migrate-cmd")

var migrations = map[string]func(*leveldb.DB) error{}

// Runner is the actor that runs the migrations
type Runner struct{}

// NewMigrationRunner creates default runner
func NewMigrationRunner() *Runner {
	return &Runner{}
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

	// For each of them, in order execute
	for _, k := range migrationList {
		start := time.Now()

		if repo.Exists(k) {
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
		mi := &Item{
			ID:       k,
			DateRun:  time.Now().UTC(),
			Duration: time.Since(start),
			Hash:     "0x", // Not implemented yet
		}
		if err = repo.CreateMigration(mi); err != nil {
			err1 := revertDBToBackup(repo, bkpRepo)
			if err1 != nil {
				return err1
			}
			return err
		}

		err = bkpRepo.Close()
		if err != nil {
			return err
		}

		log.Infof("Migration %s successfully run", k)
	}

	return repo.Close()
}

func getBackupName(path, name string) string {
	bkpPath := strings.TrimSuffix(path, ".leveldb")
	return fmt.Sprintf("%s_%s.leveldb", bkpPath, name)
}

func backupDB(srcRepo *Repository, migrationID string) (bkp *Repository, err error) {
	//Closing src to make backup
	err = srcRepo.Close()
	if err != nil {
		return nil, err
	}

	dstPath := getBackupName(srcRepo.dbPath, migrationID)
	err = CopyDir(srcRepo.dbPath, dstPath)
	if err != nil {
		return nil, err
	}

	// Refreshing src DB
	err = srcRepo.Open()
	if err != nil {
		return nil, err
	}

	return NewMigrationRepository(dstPath)
}

func revertDBToBackup(srcDB, bkpDB *Repository) error {
	err := srcDB.Close()
	if err != nil {
		return err
	}
	err = bkpDB.Close()
	if err != nil {
		return err
	}

	err = os.RemoveAll(srcDB.dbPath)
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
