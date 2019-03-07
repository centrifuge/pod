package txv1

import (
	"github.com/centrifuge/go-centrifuge/config/configstore"
	"github.com/centrifuge/go-centrifuge/storage"
	"github.com/centrifuge/go-centrifuge/transactions"
)

// Bootstrapper implements bootstrap.Bootstrapper.
type Bootstrapper struct{}

// Bootstrap adds transaction.Repository into context.
func (b Bootstrapper) Bootstrap(ctx map[string]interface{}) error {
	cfg, err := configstore.RetrieveConfig(false, ctx)
	if err != nil {
		return err
	}

	repo, ok := ctx[storage.BootstrappedDB].(storage.Repository)
	if !ok {
		return transactions.ErrTransactionBootstrap
	}

	txRepo := NewRepository(repo)
	ctx[transactions.BootstrappedRepo] = txRepo

	txSrv := NewManager(cfg, txRepo)
	ctx[transactions.BootstrappedService] = txSrv
	return nil
}
