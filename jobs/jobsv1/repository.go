package jobsv1

import (
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/jobs"
	"github.com/centrifuge/go-centrifuge/storage"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

const jobPrefix string = "job_"

// jobRepository implements Repository.
type jobRepository struct {
	repo storage.Repository
}

// NewRepository registers the the Job model and returns the an implementation
// of the Repository.
func NewRepository(repo storage.Repository) jobs.Repository {
	repo.Register(new(jobs.Job))
	return &jobRepository{repo: repo}
}

// getKey appends identity with id.
// With identity coming at first, we can even fetch jobs belonging to specific identity through prefix.
func getKey(did identity.DID, id jobs.JobID) ([]byte, error) {
	if jobs.JobIDEqual(jobs.NilJobID(), id) {
		return nil, errors.New("job ID is not valid")
	}
	hexKey := hexutil.Encode(append(did[:], id.Bytes()...))
	return append([]byte(jobPrefix), []byte(hexKey)...), nil
}

// Get returns the job associated with identity and id.
func (r *jobRepository) Get(did identity.DID, id jobs.JobID) (*jobs.Job, error) {
	key, err := getKey(did, id)
	if err != nil {
		return nil, errors.NewTypedError(jobs.ErrKeyConstructionFailed, err)
	}

	m, err := r.repo.Get(key)
	if err != nil {
		return nil, errors.NewTypedError(jobs.ErrJobsMissing, err)
	}

	return m.(*jobs.Job), nil
}

// Save saves the job to the repository.
func (r *jobRepository) Save(job *jobs.Job) error {
	key, err := getKey(job.DID, job.ID)
	if err != nil {
		return errors.NewTypedError(jobs.ErrKeyConstructionFailed, err)
	}

	if r.repo.Exists(key) {
		return r.repo.Update(key, job)
	}

	return r.repo.Create(key, job)
}
