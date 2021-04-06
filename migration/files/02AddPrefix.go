package migrationfiles

import (
	"encoding/json"
	"strings"

	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/util"
)

// value is an internal representation of how levelDb stores the model.
type value struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

// AddPrefix02 Adds db prefix to documents and jobs
func AddPrefix02(db *leveldb.DB) error {
	iter := db.NewIterator(util.BytesPrefix([]byte{}), nil)
	for iter.Next() {
		key := iter.Key()
		data := iter.Value()
		// Do nothing if entry type is prefixed already
		if isKnownPlainTextKey(key) {
			continue
		}
		v := new(value)
		err := json.Unmarshal(data, v)
		if err != nil {
			continue
		}
		prefix := []byte("document_")
		if strings.Contains(v.Type, "jobs.Job") {
			prefix = []byte("job_")
		}
		err = db.Put(append(prefix, key...), data, nil)
		if err != nil {
			return err
		}
		err = db.Delete(key, nil)
		if err != nil {
			return err
		}
	}
	iter.Release()

	err := iter.Error()
	if err != nil {
		return err
	}

	log.Infof("AddPrefix02 Migration Run successfully")
	return nil
}
