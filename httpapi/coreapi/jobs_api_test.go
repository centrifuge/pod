// +build unit

package coreapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/jobs"
	testingidentity "github.com/centrifuge/go-centrifuge/testingutils/identity"
	"github.com/centrifuge/go-centrifuge/testingutils/testingjobs"
	"github.com/go-chi/chi"
	"github.com/stretchr/testify/assert"
)

func TestService_GetJobStatus(t *testing.T) {
	getHTTPReqAndResp := func(ctx context.Context) (*httptest.ResponseRecorder, *http.Request) {
		return httptest.NewRecorder(), httptest.NewRequest("GET", "/jobs/{job_id}", nil).WithContext(ctx)
	}
	// empty job_id
	rctx := chi.NewRouteContext()
	rctx.URLParams.Keys = make([]string, 1, 1)
	rctx.URLParams.Values = make([]string, 1, 1)
	rctx.URLParams.Keys[0] = "job_id"
	rctx.URLParams.Values[0] = ""
	ctx := context.WithValue(context.Background(), chi.RouteCtxKey, rctx)
	w, r := getHTTPReqAndResp(ctx)
	h := handler{}
	h.GetJobStatus(w, r)
	assert.Equal(t, w.Code, http.StatusBadRequest)
	assert.Contains(t, w.Body.String(), ErrInvalidJobID.Error())

	// invalid jobID
	rctx.URLParams.Values[0] = "invalid value"
	w, r = getHTTPReqAndResp(ctx)
	h = handler{}
	h.GetJobStatus(w, r)
	assert.Equal(t, w.Code, http.StatusBadRequest)
	assert.Contains(t, w.Body.String(), ErrInvalidJobID.Error())

	// missing account
	jobID := jobs.NewJobID()
	rctx.URLParams.Values[0] = jobID.String()
	w, r = getHTTPReqAndResp(ctx)
	h = handler{}
	h.GetJobStatus(w, r)
	assert.Equal(t, w.Code, http.StatusNotFound)
	assert.Contains(t, w.Body.String(), ErrJobNotFound.Error())

	// missing job
	did := testingidentity.GenerateRandomDID()
	ctx = context.WithValue(ctx, config.AccountHeaderKey, did.String())
	w, r = getHTTPReqAndResp(ctx)
	jobMan := testingjobs.MockJobManager{}
	jobMan.On("GetJobStatus", did, jobID).Return(nil, errors.New("missing job"))
	h = handler{srv: Service{jobsSrv: jobMan}}
	h.GetJobStatus(w, r)
	assert.Equal(t, w.Code, http.StatusNotFound)
	assert.Contains(t, w.Body.String(), ErrJobNotFound.Error())
	jobMan.AssertExpectations(t)

	// success
	w, r = getHTTPReqAndResp(ctx)
	jobMan = testingjobs.MockJobManager{}
	tt := time.Now().UTC()
	jobMan.On("GetJobStatus", did, jobID).Return(jobs.StatusResponse{JobID: jobID.String(), LastUpdated: tt}, nil)
	h = handler{srv: Service{jobsSrv: jobMan}}
	h.GetJobStatus(w, r)
	assert.Equal(t, w.Code, http.StatusOK)
	assert.Contains(t, w.Body.String(), jobID.String())
	assert.Contains(t, w.Body.String(), tt.Format(time.RFC3339Nano))
	jobMan.AssertExpectations(t)
}
