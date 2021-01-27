package jobsv2

import (
	"bytes"
	"context"
	"sync"
	"time"

	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/gocelery/v2"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/syndtr/goleveldb/leveldb"
)

const (
	prefix                = "jobs_v2_"
	defaultReQueueTimeout = 30 * time.Minute
)

// Result represents a future result of a job
type Result interface {
	// Await blocks until job is finished to return its results.
	Await(ctx context.Context) (res interface{}, err error)
}

// Dispatcher is a task dispatcher
type Dispatcher interface {
	Name() string
	Start(ctx context.Context, wg *sync.WaitGroup, startupErr chan<- error)
	RegisterRunner(name string, runner gocelery.Runner) bool
	RegisterRunnerFunc(name string, runnerFunc gocelery.RunnerFunc) bool
	Dispatch(acc identity.DID, job *gocelery.Job) (Result, error)
	Job(acc identity.DID, jobID []byte) (*gocelery.Job, error) // TODO(ved): jobID to a defined type
	Result(acc identity.DID, jobID []byte) (Result, error)
}

type dispatcher struct {
	verifier
	*gocelery.Dispatcher
}

// NewDispatcher returns a new dispatcher with levelDB storage
func NewDispatcher(db *leveldb.DB, workerCount int, requeueTimeout time.Duration) (Dispatcher, error) {
	storage := gocelery.NewLevelDBStorage(db)
	queue := gocelery.NewQueue(storage, requeueTimeout)
	v := verifier{db: db}
	return &dispatcher{
		verifier:   v,
		Dispatcher: gocelery.NewDispatcher(workerCount, storage, queue),
	}, nil
}

func (d *dispatcher) Job(acc identity.DID, jobID []byte) (*gocelery.Job, error) {
	if !d.isJobOwner(acc, jobID) {
		return nil, gocelery.ErrNotFound
	}

	return d.Dispatcher.Job(jobID)
}

func (d *dispatcher) Dispatch(acc identity.DID, job *gocelery.Job) (Result, error) {
	// if there is a job already, error out
	if d.isJobOwner(acc, job.ID) {
		return nil, errors.New("job dispatched already")
	}

	err := d.setJobOwner(acc, job.ID)
	if err != nil {
		return nil, err
	}

	return d.Dispatcher.Dispatch(job)
}

func (d *dispatcher) Result(acc identity.DID, jobID []byte) (Result, error) {
	if !d.isJobOwner(acc, jobID) {
		return nil, gocelery.ErrNotFound
	}

	return gocelery.Result{
		JobID:      jobID,
		Dispatcher: d.Dispatcher,
	}, nil
}

func (d *dispatcher) Start(ctx context.Context, wg *sync.WaitGroup, startupErr chan<- error) {
	defer wg.Done()
	d.Dispatcher.Start(ctx)
}

func (d *dispatcher) Name() string {
	return "Jobs Dispatcher"
}

type verifier struct {
	db *leveldb.DB
}

func (v verifier) isJobOwner(acc identity.DID, jobID []byte) bool {
	key := v.getKey(acc, jobID)
	val, err := v.db.Get(key, nil)
	if err != nil {
		return false
	}

	return bytes.Equal(jobID, val)
}

func (v verifier) setJobOwner(acc identity.DID, jobID []byte) error {
	key := v.getKey(acc, jobID)
	return v.db.Put(key, jobID, nil)
}

func (v verifier) getKey(acc identity.DID, jobID []byte) []byte {
	hexKey := hexutil.Encode(append(acc[:], jobID...))
	return append([]byte(prefix), []byte(hexKey)...)
}
