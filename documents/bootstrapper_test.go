// +build unit

package documents

import (
	"testing"

	"github.com/centrifuge/go-centrifuge/config"
	"github.com/stretchr/testify/assert"
	"github.com/syndtr/goleveldb/leveldb"
)

func TestBootstrapper_Bootstrap(t *testing.T) {
	ctx := make(map[string]interface{})
	ctx[config.BootstrappedLevelDB] = &leveldb.DB{}
	err := Bootstrapper{}.Bootstrap(ctx)
	assert.Nil(t, err)
	assert.NotNil(t, ctx[BootstrappedRegistry])
	_, ok := ctx[BootstrappedRegistry].(*ServiceRegistry)
	assert.True(t, ok)
}
