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

type content struct {
	Name string `json:"name"`
}

func TestIsHexKey(t *testing.T) {
	b := migrationutils.RandomSlice(10)
	assert.False(t, isHexKey(b))

	b = []byte{48}
	assert.False(t, isHexKey(b))

	b = []byte{48, 48, 50, 52}
	assert.True(t, isHexKey(b))
}

func TestIsKnownPlainTextKey(t *testing.T) {
	input := "config"
	assert.True(t, isKnownPlainTextKey([]byte(input)))
	input = "migration_123456"
	assert.True(t, isKnownPlainTextKey([]byte(input)))
	input = "account-o1231234"
	assert.True(t, isKnownPlainTextKey([]byte(input)))
	input = string(migrationutils.RandomSlice(12))
	assert.False(t, isKnownPlainTextKey([]byte(input)))
}

func TestKeysToHex01(t *testing.T) {
	prefix := fmt.Sprintf("/tmp/datadir_%x", migrationutils.RandomByte32())
	targetDir := fmt.Sprintf("%s.leveldb", prefix)

	// Cleanup after test
	defer migrationutils.CleanupDBFiles(prefix)

	db, err := leveldb.OpenFile(targetDir, nil)
	assert.NoError(t, err)

	var keys [][]byte
	// Add some non-hex keys
	for i := 0; i < 5; i++ {
		key := migrationutils.RandomSlice(32)
		keys = append(keys, key)
		data, err := json.Marshal(&content{fmt.Sprintf("Val_%d", i)})
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

	err = KeysToHex01(db)
	assert.NoError(t, err)

	for i := 0; i < 5; i++ {
		// non-hex key not found
		_, err = db.Get(keys[i], nil)
		assert.Error(t, err)

		v, err := db.Get([]byte(hexutil.Encode(keys[i])), nil)
		assert.NoError(t, err)
		var c content
		err = json.Unmarshal(v, &c)
		assert.NoError(t, err)
		assert.Equal(t, fmt.Sprintf("Val_%d", i), c.Name)
	}

	_, err = db.Get([]byte("migration_123143"), nil)
	assert.NoError(t, err)
	_, err = db.Get([]byte("config"), nil)
	assert.NoError(t, err)
	_, err = db.Get([]byte("account-123143"), nil)
	assert.NoError(t, err)

}
