package jobs

import (
	"context"
	"fmt"
	"sync"

	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/storage/leveldb"
	ldb "github.com/syndtr/goleveldb/leveldb"
)

// BootstrappedJobDispatcher is a key to access dispatcher
const BootstrappedJobDispatcher = "BootstrappedJobDispatcher"

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

	ctx[BootstrappedJobDispatcher] = d
	return nil
}

var (
	dispatcherCtx       context.Context
	dispatcherCtxCanc   context.CancelFunc
	dispatcherWaitGroup sync.WaitGroup
)

func (b Bootstrapper) TestBootstrap(ctx map[string]interface{}) error {
	if err := b.Bootstrap(ctx); err != nil {
		return err
	}

	dispatcher := ctx[BootstrappedJobDispatcher].(Dispatcher)

	dispatcherCtx, dispatcherCtxCanc = context.WithCancel(context.Background())

	go dispatcher.Start(dispatcherCtx, &dispatcherWaitGroup, nil)

	return nil
}

func (b Bootstrapper) TestTearDown() error {
	dispatcherCtxCanc()
	dispatcherWaitGroup.Wait()

	return nil
}
