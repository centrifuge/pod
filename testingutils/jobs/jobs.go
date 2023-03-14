//go:build unit || integration || testworld

package jobs

import (
	"context"
	"fmt"
	"time"

	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/centrifuge/gocelery/v2"
	"github.com/centrifuge/pod/jobs"
)

const (
	jobTimeout      = 3 * time.Minute
	jobPollInterval = 1 * time.Second
)

func WaitForJobToFinish(
	ctx context.Context,
	dispatcher jobs.Dispatcher,
	identity *types.AccountID,
	jobID gocelery.JobID,
) error {
	ctx, cancel := context.WithTimeout(ctx, jobTimeout)
	defer cancel()

	t := time.NewTicker(jobPollInterval)
	defer t.Stop()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("context done while waiting for job to finish: %w", ctx.Err())
		case <-t.C:
			job, err := dispatcher.Job(identity, jobID)

			if err != nil {
				return err
			}

			if job.HasCompleted() {
				return nil
			}
		}
	}
}
