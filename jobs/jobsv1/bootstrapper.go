package jobsv1

import (
	"github.com/centrifuge/go-centrifuge/config/configstore"
	"github.com/centrifuge/go-centrifuge/jobs"
	"github.com/centrifuge/go-centrifuge/storage"
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
		return jobs.ErrJobsBootstrap
	}

	jobsRepo := NewRepository(repo)
	ctx[jobs.BootstrappedRepo] = jobsRepo

	jobsMan := NewManager(cfg, jobsRepo)
	ctx[jobs.BootstrappedService] = jobsMan
	return nil
}
