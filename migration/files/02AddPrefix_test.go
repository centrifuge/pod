package migrationfiles

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/centrifuge/go-centrifuge/migration/utils"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/stretchr/testify/assert"
	"github.com/syndtr/goleveldb/leveldb"
)

func TestAddPrefix02(t *testing.T) {
	prefix := fmt.Sprintf("/tmp/datadir_%x", migrationutils.RandomByte32())
	targetDir := fmt.Sprintf("%s.leveldb", prefix)

	// Cleanup after test
	defer migrationutils.CleanupDBFiles(prefix)

	db, err := leveldb.OpenFile(targetDir, nil)
	assert.NoError(t, err)

	var keys [][]byte
	// Add some non-hex keys
	for i := 0; i < 5; i++ {
		key := []byte(hexutil.Encode(migrationutils.RandomSlice(32)))
		keys = append(keys, key)
		innerType := "invoice.Invoice"
		if i%2 != 0 {
			innerType = "jobs.Job"
		}
		data, err := json.Marshal(&value{Type: innerType, Data: json.RawMessage([]byte(fmt.Sprintf("{\"val\":\"Val_%d\"}", i)))})
		assert.NoError(t, err)
		err = db.Put(key, data, nil)
		assert.NoError(t, err)
	}

	// Add plain text keys
	data, err := json.Marshal(&content{"Val_Migration"})
	assert.NoError(t, err)
	err = db.Put([]byte("migration_123143"), data, nil)
	assert.NoError(t, err)
	data, err = json.Marshal(&content{"Val_config"})
	assert.NoError(t, err)
	err = db.Put([]byte("config"), data, nil)
	assert.NoError(t, err)
	data, err = json.Marshal(&content{"Val_Account"})
	assert.NoError(t, err)
	err = db.Put([]byte("account-123143"), data, nil)
	assert.NoError(t, err)

	err = AddPrefix02(db)
	assert.NoError(t, err)

	for i := 0; i < 5; i++ {
		// non-prefixed key not found
		_, err = db.Get(keys[i], nil)
		assert.Error(t, err)
		innerType := "document_"
		if i%2 != 0 {
			innerType = "job_"
		}
		v, err := db.Get(append([]byte(innerType), keys[i]...), nil)
		assert.NoError(t, err)
		var c value
		err = json.Unmarshal(v, &c)
		assert.NoError(t, err)
		cData, err := c.Data.MarshalJSON()
		assert.NoError(t, err)
		assert.Equal(t, []byte(fmt.Sprintf("{\"val\":\"Val_%d\"}", i)), cData)
	}

	_, err = db.Get([]byte("migration_123143"), nil)
	assert.NoError(t, err)
	_, err = db.Get([]byte("config"), nil)
	assert.NoError(t, err)
	_, err = db.Get([]byte("account-123143"), nil)
	assert.NoError(t, err)

}
