package accounts

import (
	"errors"

	"github.com/centrifuge/go-centrifuge/storage"
)

const (
	// BootstrappedAccountSrv is a key mapped to Account Service at boot
	BootstrappedAccountSrv string = "BootstrappedAccountSrv"
)

// Bootstrapper implements bootstrapper.Bootstrapper.
type Bootstrapper struct{}

// Bootstrap initialises the accounts service.
func (Bootstrapper) Bootstrap(context map[string]interface{}) error {
	db, ok := context[storage.BootstrappedDB].(storage.Repository)
	if !ok {
		return errors.New("storage repository not initialised")
	}

	repo := newRepository(db)
	srv := newService(repo)
	context[BootstrappedAccountSrv] = srv
	return nil
}
