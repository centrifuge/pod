package jobs

import (
	"fmt"

	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/storage/leveldb"
	ldb "github.com/syndtr/goleveldb/leveldb"
)

// BootstrappedDispatcher is a key to access dispatcher
const BootstrappedDispatcher = "BootstrappedDispatcher"

// Bootstrapper implements bootstrap.Bootstrapper.
type Bootstrapper struct{}

// Bootstrap adds transaction.Repository into context.
func (b Bootstrapper) Bootstrap(ctx map[string]interface{}) error {
	db := ctx[leveldb.BootstrappedLevelDB].(*ldb.DB)
	cfg, err := config.RetrieveConfig(false, ctx)
	if err != nil {
		return err
	}

	d, err := NewDispatcher(db, cfg.GetNumWorkers(), defaultReQueueTimeout)
	if err != nil {
		return fmt.Errorf("failed to init dispatcher: %w", err)
	}

	ctx[BootstrappedDispatcher] = d
	return nil
}
