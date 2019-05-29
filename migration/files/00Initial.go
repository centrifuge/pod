package migrationfiles

import (
	logging "github.com/ipfs/go-log"
	"github.com/syndtr/goleveldb/leveldb"
)

var log = logging.Logger("migrate-files")

// Initial00 Does nothing
func Initial00(db *leveldb.DB) error {
	log.Infof("00Initial Migration Run successfully")
	return nil
}
