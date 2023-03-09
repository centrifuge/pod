package documents

import (
	"context"
	"encoding/gob"
	"fmt"
	"time"

	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/centrifuge/gocelery/v2"
	"github.com/centrifuge/pod/config"
	"github.com/centrifuge/pod/contextutil"
	"github.com/centrifuge/pod/jobs"
)

func init() {
	gob.Register(&types.AccountID{})
}

type task struct {
	runnerFunc gocelery.RunnerFunc
	next       string
}

const anchorJob = "Commit document anchor"

// AnchorJob is a async anchoring task
// args should be as follows
// DID, versionID, preCommit(true|false)
// ignores overrides
type AnchorJob struct {
	configSrv config.Service
	repo      Repository
	processor AnchorProcessor

	tasks map[string]task
}

// New returns a new instance of anchor Job
func (a *AnchorJob) New() gocelery.Runner {
	aj := &AnchorJob{
		configSrv: a.configSrv,
		repo:      a.repo,
		processor: a.processor,
	}
	aj.loadTasks()
	return aj
}

// RunnerFunc will return a RunnerFunc for a given task
func (a *AnchorJob) RunnerFunc(task string) gocelery.RunnerFunc {
	return a.tasks[task].runnerFunc
}

// Next returns the next task in line after task provided
func (a *AnchorJob) Next(task string) (next string, ok bool) {
	t := a.tasks[task]
	return t.next, t.next != ""
}

func (a *AnchorJob) runnerFunc(run func(ctx context.Context, doc Document) error) gocelery.RunnerFunc {
	return func(args []interface{}, overrides map[string]interface{}) (result interface{}, err error) {
		identity := args[0].(*types.AccountID)
		versionID := args[1].([]byte)
		doc, err := a.repo.Get(identity.ToBytes(), versionID)
		if err != nil {
			return nil, fmt.Errorf("failed to get document from ID and Version: %w", err)
		}

		acc, err := a.configSrv.GetAccount(identity.ToBytes())
		if err != nil {
			return nil, fmt.Errorf("failed to get account from config service: %w", err)
		}

		ctx := contextutil.WithAccount(context.Background(), acc)

		err = run(ctx, doc)
		if err != nil {
			return nil, err
		}

		return nil, a.repo.Update(identity.ToBytes(), versionID, doc)
	}
}

func (a *AnchorJob) loadTasks() {
	a.tasks = map[string]task{
		"prepare_request_signatures": {
			runnerFunc: a.runnerFunc(a.processor.PrepareForSignatureRequests),
			next:       "pre_commit",
		},
		"pre_commit": {
			runnerFunc: func(
				args []interface{},
				overrides map[string]interface{},
			) (interface{}, error) {
				preCommit := args[2].(bool)
				if !preCommit {
					return nil, nil
				}

				return a.runnerFunc(a.processor.PreAnchorDocument)(args, overrides)
			},
			next: "request_signatures",
		},
		"request_signatures": {
			runnerFunc: a.runnerFunc(a.processor.RequestSignatures),
			next:       "prepare_anchor",
		},
		"prepare_anchor": {
			runnerFunc: a.runnerFunc(a.processor.PrepareForAnchoring),
			next:       "anchor_document",
		},
		"anchor_document": {
			runnerFunc: a.runnerFunc(a.processor.AnchorDocument),
			next:       "set_document_committed",
		},
		"set_document_committed": {
			runnerFunc: a.runnerFunc(func(ctx context.Context, doc Document) error {
				return doc.SetStatus(Committed)
			}),
			next: "send_document",
		},
		"send_document": {
			runnerFunc: a.runnerFunc(a.processor.SendDocument),
		},
	}
}

// initiateAnchorJob initiate document anchor job
func initiateAnchorJob(
	dispatcher jobs.Dispatcher,
	accountID *types.AccountID,
	versionID []byte,
	preCommit bool,
) (jobID gocelery.JobID, err error) {
	job := gocelery.NewRunnerJob(
		"Document anchor commit job",
		anchorJob,
		"prepare_request_signatures",
		[]interface{}{accountID, versionID, preCommit},
		make(map[string]interface{}),
		time.Time{},
	)

	_, err = dispatcher.Dispatch(accountID, job)
	if err != nil {
		return nil, err
	}

	return job.ID, nil
}
