package jobs

import "github.com/centrifuge/go-centrifuge/errors"

const (

	// ErrJobsBootstrap error when bootstrap fails.
	ErrJobsBootstrap = errors.Error("failed to bootstrap jobs")

	// ErrJobsMissing error when job doesn't exist in Repository.
	ErrJobsMissing = errors.Error("job doesn't exist")

	// ErrKeyConstructionFailed error when the key construction failed.
	ErrKeyConstructionFailed = errors.Error("failed to construct job key")
)
