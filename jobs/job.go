package jobs

import (
	"context"
	"encoding/json"
	"reflect"
	"time"

	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/satori/go.uuid"
)

// Status represents the status of the job
type Status string

const (
	// Status constants

	// Success is the success status for a job or a task
	Success Status = "success"
	// Failed is the failed status for a job or a task
	Failed Status = "failed"
	// Pending is the pending status for a job or a task
	Pending Status = "pending"

	// JobIDParam maps job ID in the kwargs.
	JobIDParam = "jobID"

	// BootstrappedRepo is the key mapped to jobs.Repository.
	BootstrappedRepo = "BootstrappedRepo"

	// BootstrappedService is the key to mapped jobs.JobManager
	BootstrappedService = "BootstrappedService"

	// JobDataTypeURL is the type of the job data
	JobDataTypeURL = "http://github.com/centrifuge/go-centrifuge/jobs/#Job"
)

// Log represents a single task in a job.
type Log struct {
	Action    string
	Message   string
	CreatedAt time.Time
}

// NewLog constructs a new log with action and message
func NewLog(action, message string) Log {
	return Log{
		Action:    action,
		Message:   message,
		CreatedAt: time.Now().UTC(),
	}
}

// JobID is a centrifuge job ID. Internally represented by a UUID. Externally visible as a byte slice or a hex encoded string.
type JobID uuid.UUID

// NewJobID creates a new JobID
func NewJobID() JobID {
	u := uuid.Must(uuid.NewV4())
	return JobID(u)
}

// FromString tries to convert the given hex string jobID into a type JobID
func FromString(jobIDHex string) (JobID, error) {
	bytes, err := hexutil.Decode(jobIDHex)
	if err != nil {
		return NilJobID(), err
	}

	u, err := uuid.FromBytes(bytes)
	if err != nil {
		return NilJobID(), err
	}

	return JobID(u), nil
}

// NilJobID returns a nil JobID
func NilJobID() JobID {
	return JobID(uuid.Nil)
}

// String marshals a JobID to its hex string form
func (t JobID) String() string {
	if uuid.UUID(t) == uuid.Nil {
		return ""
	}

	return hexutil.Encode(t[:])
}

// Bytes returns the byte slice representation of the JobID
func (t JobID) Bytes() []byte {
	return uuid.UUID(t).Bytes()
}

// JobIDEqual checks if given two JobIDs are equal
func JobIDEqual(t1 JobID, t2 JobID) bool {
	u1 := uuid.UUID(t1)
	u2 := uuid.UUID(t2)
	return uuid.Equal(u1, u2)
}

// Job contains details of Job.
type Job struct {
	ID          JobID
	DID         identity.DID
	Description string

	// Status is the overall status of the Job
	Status Status

	// TaskStatus tracks the status of individual tasks running in the system for this Job
	TaskStatus map[string]Status

	// Logs are Job log messages
	Logs      []Log
	CreatedAt time.Time

	// Values retrieved from events
	Values map[string]JobValue
}

// JSON returns json marshaled job.
func (t *Job) JSON() ([]byte, error) {
	return json.Marshal(t)
}

// FromJSON loads the data into job.
func (t *Job) FromJSON(data []byte) error {
	return json.Unmarshal(data, t)
}

// Type returns the reflect.Type of the job.
func (t *Job) Type() reflect.Type {
	return reflect.TypeOf(t)
}

// NewJob returns a new Job with a pending state
func NewJob(identity identity.DID, description string) *Job {
	return &Job{
		ID:          NewJobID(),
		DID:         identity,
		Description: description,
		Status:      Pending,
		TaskStatus:  make(map[string]Status),
		CreatedAt:   time.Now().UTC(),
		Values:      make(map[string]JobValue),
	}
}

// JobValue holds the key and value filtered by the Job
type JobValue struct {
	Key    string
	KeyIdx int
	Value  []byte
}

// StatusResponse holds the job status details.
type StatusResponse struct {
	JobID       string    `json:"job_id"`
	Status      string    `json:"status"`
	Message     string    `json:"message"`
	LastUpdated time.Time `json:"last_updated" swaggertype:"primitive,string"`
}

// Config is the config interface for jobs package
type Config interface {
	GetEthereumContextWaitTimeout() time.Duration
}

// Manager is a manager for centrifuge Jobs.
type Manager interface {
	// ExecuteWithinJob executes the given unit of work within a Job
	ExecuteWithinJob(ctx context.Context, accountID identity.DID, existingJobID JobID, desc string, work func(accountID identity.DID, jobID JobID, jobManager Manager, err chan<- error)) (jobID JobID, done chan error, err error)
	GetJob(accountID identity.DID, id JobID) (*Job, error)
	UpdateJobWithValue(accountID identity.DID, id JobID, key string, value []byte) error
	UpdateTaskStatus(accountID identity.DID, id JobID, status Status, taskName, message string) error
	GetJobStatus(accountID identity.DID, id JobID) (StatusResponse, error)
	WaitForJob(accountID identity.DID, txID JobID) error
	GetDefaultTaskTimeout() time.Duration
}

// Repository can be implemented by a type that handles storage for Jobs.
type Repository interface {
	Get(did identity.DID, id JobID) (*Job, error)
	Save(job *Job) error
}
