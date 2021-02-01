package documents

import (
	"context"
	"encoding/gob"
	"fmt"

	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/gocelery/v2"
)

func init() {
	gob.Register(identity.DID{})
}

type task struct {
	runnerFunc gocelery.RunnerFunc
	next       string
}

// AnchorJob is a async anchoring task
// args should be as follows
// DID, versionID, preCommit(true|false)
// ignores overrides
type AnchorJob struct {
	repo      Repository
	processor AnchorProcessor

	tasks map[string]task
}

// New returns a new instance of anchor Job
func (a AnchorJob) New() gocelery.Runner {
	na := AnchorJob{processor: a.processor}
	return na
}

// RunnerFunc will return a RunnerFunc for a given task
func (a AnchorJob) RunnerFunc(task string) gocelery.RunnerFunc {
	return a.tasks[task].runnerFunc
}

// Next returns the next task in line after task provided
func (a AnchorJob) Next(task string) (next string, ok bool) {
	t := a.tasks[task]
	return t.next, t.next != ""
}

func (a AnchorJob) runnerFunc(run func(ctx context.Context, doc Document) error) gocelery.RunnerFunc {
	return func(args []interface{}, overrides map[string]interface{}) (result interface{}, err error) {
		did := args[0].(identity.DID)
		versionID := args[1].([]byte)
		doc, err := a.repo.Get(did[:], versionID)
		if err != nil {
			return nil, fmt.Errorf("failed to get document from ID and Version: %w", err)
		}

		err = run(context.Background(), doc)
		if err != nil {
			return nil, err
		}

		return nil, a.repo.Update(did[:], versionID, doc)
	}
}

func (a AnchorJob) loadTasks() {
	tasks := make(map[string]task)
	tasks["prepare_request_signatures"] = task{
		runnerFunc: a.runnerFunc(a.processor.PrepareForSignatureRequests),
		next:       "pre_commit"}
	tasks["pre_commit"] = task{
		runnerFunc: func(args []interface{}, overrides map[string]interface{}) (interface{},
			error) {
			preCommit := args[2].(bool)
			if !preCommit {
				return nil, nil
			}

			return a.runnerFunc(a.processor.PreAnchorDocument)(args, overrides)
		},
		next: "request_signatures"}
	tasks["request_signatures"] = task{
		runnerFunc: a.runnerFunc(a.processor.RequestSignatures),
		next:       "prepare_anchor"}
	tasks["prepare_anchor"] = task{
		runnerFunc: a.runnerFunc(a.processor.PrepareForAnchoring),
		next:       "anchor_document"}
	tasks["anchor_document"] = task{
		runnerFunc: a.runnerFunc(a.processor.AnchorDocument),
		next:       "send_document",
	}
	tasks["send_document"] = task{runnerFunc: a.runnerFunc(a.processor.SendDocument)}
	a.tasks = tasks
}
