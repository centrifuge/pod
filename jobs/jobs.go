package jobs

import (
	"bytes"
	"context"
	"encoding/gob"
	"fmt"
	"sync"
	"time"

	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/centrifuge/gocelery/v2"
	"github.com/centrifuge/pod/bootstrap"
	"github.com/centrifuge/pod/config"
	"github.com/centrifuge/pod/contextutil"
	"github.com/centrifuge/pod/errors"
	"github.com/centrifuge/pod/notification"
	"github.com/centrifuge/pod/utils/byteutils"
	"github.com/ethereum/go-ethereum/common/hexutil"
	logging "github.com/ipfs/go-log"
	"github.com/syndtr/goleveldb/leveldb"
)

func init() {
	gob.Register(gocelery.JobID{})
}

const (
	prefix                = "jobs_v2_"
	defaultReQueueTimeout = 30 * time.Minute
)

var log = logging.Logger("jobs-dispatcher")

//go:generate mockery --name Result --structname ResultMock --filename result_mock.go --inpackage

// Result represents a future result of a job
type Result interface {
	// Await blocks until job is finished to return its results.
	Await(ctx context.Context) (res interface{}, err error)
}

//go:generate mockery --name Dispatcher --structname DispatcherMock --filename dispatcher_mock.go --inpackage

// Dispatcher is a task dispatcher
type Dispatcher interface {
	Name() string
	Start(ctx context.Context, wg *sync.WaitGroup, startupErr chan<- error)
	RegisterRunner(name string, runner gocelery.Runner) bool
	RegisterRunnerFunc(name string, runnerFunc gocelery.RunnerFunc) bool
	Dispatch(accountID *types.AccountID, job *gocelery.Job) (Result, error)
	Job(accountID *types.AccountID, jobID gocelery.JobID) (*gocelery.Job, error)
	Result(accountID *types.AccountID, jobID gocelery.JobID) (Result, error)
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

func (d *dispatcher) Job(accountID *types.AccountID, jobID gocelery.JobID) (*gocelery.Job, error) {
	if !d.isJobOwner(accountID, jobID) {
		return nil, gocelery.ErrNotFound
	}

	return d.Dispatcher.Job(jobID)
}

func (d *dispatcher) Dispatch(accountID *types.AccountID, job *gocelery.Job) (Result, error) {
	// if there is a job already, error out
	if d.isJobOwner(accountID, job.ID) {
		return nil, errors.New("job dispatched already")
	}

	err := d.setJobOwner(accountID, job.ID)
	if err != nil {
		return nil, err
	}

	return d.Dispatcher.Dispatch(job)
}

func (d *dispatcher) Result(accountID *types.AccountID, jobID gocelery.JobID) (Result, error) {
	if !d.isJobOwner(accountID, jobID) {
		return nil, gocelery.ErrNotFound
	}

	return gocelery.Result{
		JobID:      jobID,
		Dispatcher: d.Dispatcher,
	}, nil
}

func (d *dispatcher) Start(ctx context.Context, wg *sync.WaitGroup, startupErr chan<- error) {
	// start job finished notifier
	wg.Add(1)

	go func() {
		if err := initJobWebhooks(ctx, d, wg); err != nil {
			startupErr <- err
		}
	}()

	// start dispatcher
	defer wg.Done()
	d.Dispatcher.Start(ctx)
}

func (d *dispatcher) Name() string {
	return "Jobs Dispatcher"
}

func initJobWebhooks(ctx context.Context, dispatcher *dispatcher, wg *sync.WaitGroup) error {
	defer wg.Done()

	cctx, ok := ctx.Value(bootstrap.NodeObjRegistry).(map[string]interface{})
	if !ok {
		log.Error("jobs: failed to find Node registry")
		return errors.New("node registry not found")
	}

	configSrv, ok := cctx[config.BootstrappedConfigStorage].(config.Service)
	if !ok {
		log.Error("jobs: failed to find config service")
		return errors.New("config service not found")
	}

	sender := notification.NewWebhookSender()
	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("context done while running job webhooks: %w", ctx.Err())
		case job := <-dispatcher.OnFinished():
			owner, err := dispatcher.jobOwner(job.ID)
			if err != nil {
				log.Errorf("failed to get owner for the job[%v]: %v", job.ID, err)
				continue
			}

			acc, err := configSrv.GetAccount(owner.ToBytes())
			if err != nil {
				log.Errorf("failed to find account for the job[%v]: %v", job.ID, err)
				continue
			}

			message := notification.Message{
				EventType:  notification.EventTypeJob,
				RecordedAt: time.Now().UTC(),
				Job: &notification.JobMessage{
					ID:         byteutils.HexBytes(job.ID),
					Desc:       job.Desc,
					ValidUntil: job.ValidUntil,
					FinishedAt: job.FinishedAt,
				},
			}

			err = sender.Send(contextutil.WithAccount(ctx, acc), message)
			if err != nil {
				log.Errorf("failed to send job message: %v", err)
			}
		}
	}
}

type verifier struct {
	db *leveldb.DB
}

func (v verifier) isJobOwner(accountID *types.AccountID, jobID []byte) bool {
	key := v.getKey(jobID)
	val, err := v.db.Get(key, nil)
	if err != nil {
		return false
	}

	return bytes.Equal(accountID[:], val)
}

func (v verifier) setJobOwner(accountID *types.AccountID, jobID []byte) error {
	key := v.getKey(jobID)
	return v.db.Put(key, accountID[:], nil)
}

func (v verifier) getKey(jobID []byte) []byte {
	return append([]byte(prefix), []byte(hexutil.Encode(jobID))...)
}

func (v verifier) jobOwner(jobID []byte) (*types.AccountID, error) {
	key := v.getKey(jobID)
	val, err := v.db.Get(key, nil)
	if err != nil {
		return nil, gocelery.ErrNotFound
	}

	return types.NewAccountID(val)
}
