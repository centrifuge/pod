package migrationfiles

import (
	"encoding/hex"
	"regexp"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/util"
)

// KeysToHex01 Converts all keys to hex
func KeysToHex01(db *leveldb.DB) error {
	iter := db.NewIterator(util.BytesPrefix([]byte{}), nil)
	for iter.Next() {
		key := iter.Key()
		data := iter.Value()
		// Do nothing if key is already hex
		if isHexKey(key) || isKnownPlainTextKey(key) {
			continue
		}
		err := db.Put([]byte(hexutil.Encode(key)), data, nil)
		if err != nil {
			return err
		}
		err = db.Delete(key, nil)
		if err != nil {
			return err
		}
	}
	iter.Release()

	log.Infof("01KeysToHex Migration Run successfully")
	return iter.Error()
}

func isHexKey(key []byte) bool {
	dst := make([]byte, hex.DecodedLen(len(key)))
	_, err := hex.Decode(dst, key)
	return err == nil
}

func isKnownPlainTextKey(key []byte) bool {
	elem := regexp.MustCompile("migration_(.)*")
	mf := elem.FindAllString(string(key), 1)
	elem = regexp.MustCompile("account-(.)*")
	af := elem.FindAllString(string(key), 1)
	elem = regexp.MustCompile("document_(.)*")
	df := elem.FindAllString(string(key), 1)

	return string(key) == "config" || len(mf) > 0 || len(af) > 0 || len(df) > 0
}
