package pending

import (
	"github.com/centrifuge/pod/documents"
	"github.com/centrifuge/pod/errors"
	"github.com/centrifuge/pod/storage"
)

const (
	// BootstrappedPendingDocumentService is the key to bootstrapped document service
	BootstrappedPendingDocumentService = "BootstrappedPendingDocumentService"
)

// Bootstrapper implements bootstrap.Bootstrapper.
type Bootstrapper struct{}

// Bootstrap sets the required storage and registers
func (Bootstrapper) Bootstrap(ctx map[string]interface{}) error {
	docSrv, ok := ctx[documents.BootstrappedDocumentService].(documents.Service)
	if !ok {
		return errors.New("%s not found in the bootstrapper", documents.BootstrappedDocumentService)
	}

	ldb, ok := ctx[storage.BootstrappedDB].(storage.Repository)
	if !ok {
		return errors.New("%s not found in the bootstrapper", storage.BootstrappedDB)
	}
	repo := NewRepository(ldb)
	ctx[BootstrappedPendingDocumentService] = NewService(docSrv, repo)
	return nil
}
