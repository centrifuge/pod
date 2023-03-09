package jobs

import (
	"context"
	"fmt"
	"sync"

	"github.com/centrifuge/pod/config"
	"github.com/centrifuge/pod/storage/leveldb"
	ldb "github.com/syndtr/goleveldb/leveldb"
)

// BootstrappedJobDispatcher is a key to access dispatcher
const BootstrappedJobDispatcher = "BootstrappedJobDispatcher"

// Bootstrapper implements bootstrap.Bootstrapper.
type Bootstrapper struct {
	testDispatcherCtx       context.Context
	testDispatcherCtxCanc   context.CancelFunc
	testDispatcherWaitGroup sync.WaitGroup
}

// Bootstrap adds transaction.Repository into context.
func (b *Bootstrapper) Bootstrap(ctx map[string]interface{}) error {
	db := ctx[leveldb.BootstrappedLevelDB].(*ldb.DB)
	cfg, err := config.RetrieveConfig(false, ctx)
	if err != nil {
		return err
	}

	d, err := NewDispatcher(db, cfg.GetNumWorkers(), defaultReQueueTimeout)
	if err != nil {
		return fmt.Errorf("failed to init dispatcher: %w", err)
	}

	ctx[BootstrappedJobDispatcher] = d
	return nil
}
