package v2

import (
	"net/http"

	"github.com/centrifuge/go-centrifuge/contextutil"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/utils/httputils"
	"github.com/centrifuge/gocelery/v2"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/go-chi/chi"
	"github.com/go-chi/render"
)

const (
	// ErrInvalidJobID is the sentinel error when the job_id passed is invalid.
	ErrInvalidJobID = errors.Error("Invalid Job ID")

	// ErrJobNotFound is a sentinel error when job associated with job_id is not found.
	ErrJobNotFound = errors.Error("Job not found")

	jobIDParam = "job_id"
)

// Job is an alias for gocelery Job for swagger generation
type Job = gocelery.Job

// Job returns the details of a given job.
// @summary Returns the details of a given Job.
// @description Returns the details of a given Job.
// @id get_job
// @tags Jobs
// @param authorization header string true "Hex encoded centrifuge ID of the account for the intended API action"
// @param job_id path string true "Hex encoded Job ID"
// @produce json
// @Failure 403 {object} httputils.HTTPError
// @Failure 400 {object} httputils.HTTPError
// @Failure 404 {object} httputils.HTTPError
// @success 200 {object} v2.Job
// @router /v2/jobs/{job_id} [get]
func (h handler) Job(w http.ResponseWriter, r *http.Request) {
	var err error
	var code int
	defer httputils.RespondIfError(&code, &err, w, r)

	jobID, err := hexutil.Decode(chi.URLParam(r, jobIDParam))
	if err != nil {
		err = errors.NewTypedError(ErrInvalidJobID, err)
		code = http.StatusBadRequest
		log.Error(err)
		return
	}

	account, err := contextutil.DIDFromContext(r.Context())
	if err != nil {
		log.Error(err)
		err = ErrJobNotFound
		code = http.StatusNotFound
		return
	}

	resp, err := h.srv.Job(account, jobID)
	if err != nil {
		log.Error(err)
		err = ErrJobNotFound
		code = http.StatusNotFound
		return
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, resp)
}
