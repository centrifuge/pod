// +build unit integration

package jobsv1

import (
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/jobs"
)

func (b Bootstrapper) TestBootstrap(ctx map[string]interface{}) error {
	return b.Bootstrap(ctx)
}

func (b Bootstrapper) TestTearDown() error {
	return nil
}

// extendedManager exposes package specific functions.
type extendedManager interface {
	jobs.Manager

	// saveJob only exposed for testing within package.
	// DO NOT use this outside of the package, use ExecuteWithinJob to initiate a transaction with management.
	saveJob(job *jobs.Job) error

	// createJob only exposed for testing within package.
	// DO NOT use this outside of the package, use ExecuteWithinJob to initiate a job with management.
	createJob(accountID identity.DID, desc string) (*jobs.Job, error)
}
