package coreapi

import (
	"net/http"

	"github.com/centrifuge/go-centrifuge/contextutil"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/jobs"
	"github.com/centrifuge/go-centrifuge/utils/httputils"
	"github.com/go-chi/chi"
	"github.com/go-chi/render"
)

const (
	// ErrInvalidJobID is the sentinel error when the job_id passed is invalid.
	ErrInvalidJobID = errors.Error("Invalid Job ID")

	// ErrJobNotFound is a sentinel error when job associated with job_id is not found.
	ErrJobNotFound = errors.Error("Job not found")
)

// GetJobStatus returns the status of a given job.
// @summary Returns the status of a given Job.
// @description Returns the status of a given Job.
// @id get_job_status
// @tags Jobs
// @param authorization header string true "Hex encoded centrifuge ID of the account for the intended API action"
// @param job_id path string true "Job ID"
// @produce json
// @Failure 403 {object} httputils.HTTPError
// @Failure 400 {object} httputils.HTTPError
// @Failure 404 {object} httputils.HTTPError
// @success 200 {object} jobs.StatusResponse
// @router /v1/jobs/{job_id} [get]
func (h handler) GetJobStatus(w http.ResponseWriter, r *http.Request) {
	var err error
	var code int
	defer httputils.RespondIfError(&code, &err, w, r)

	jobID, err := jobs.FromString(chi.URLParam(r, jobIDParam))
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

	resp, err := h.srv.GetJobStatus(account, jobID)
	if err != nil {
		log.Error(err)
		err = ErrJobNotFound
		code = http.StatusNotFound
		return
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, resp)
}
