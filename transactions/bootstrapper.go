package transactions

import (
	"github.com/centrifuge/go-centrifuge/config/configstore"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/storage"
)

const (
	// ErrTransactionBootstrap error when bootstrap fails.
	ErrTransactionBootstrap = errors.Error("failed to bootstrap transactions")

	// BootstrappedRepo is the key mapped to transactions.Repository.
	BootstrappedRepo = "BootstrappedRepo"

	// BootstrappedService is the key to mapped transactions.Manager
	BootstrappedService = "BootstrappedService"
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
		return ErrTransactionBootstrap
	}

	txRepo := NewRepository(repo)
	ctx[BootstrappedRepo] = txRepo

	txSrv := NewManager(cfg, txRepo)
	ctx[BootstrappedService] = txSrv
	return nil
}
