//go:build unit
// +build unit

package v2

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/jobs"

	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/centrifuge/gocelery/v2"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/go-chi/chi"
	"github.com/stretchr/testify/assert"
)

func TestService_GetJob(t *testing.T) {
	getHTTPReqAndResp := func(ctx context.Context) (*httptest.ResponseRecorder, *http.Request) {
		return httptest.NewRecorder(), httptest.NewRequest("GET", "/jobs/{job_id}", nil).WithContext(ctx)
	}
	// empty job_id
	rctx := chi.NewRouteContext()
	rctx.URLParams.Keys = make([]string, 1)
	rctx.URLParams.Values = make([]string, 1)
	rctx.URLParams.Keys[0] = "job_id"
	rctx.URLParams.Values[0] = ""
	ctx := context.WithValue(context.Background(), chi.RouteCtxKey, rctx)
	w, r := getHTTPReqAndResp(ctx)
	h := handler{}
	h.Job(w, r)
	assert.Equal(t, w.Code, http.StatusBadRequest)
	assert.Contains(t, w.Body.String(), ErrInvalidJobID.Error())

	// invalid jobID
	rctx.URLParams.Values[0] = "invalid value"
	w, r = getHTTPReqAndResp(ctx)
	h = handler{}
	h.Job(w, r)
	assert.Equal(t, w.Code, http.StatusBadRequest)
	assert.Contains(t, w.Body.String(), ErrInvalidJobID.Error())

	// missing account
	jobID := gocelery.JobID(utils.RandomSlice(32))
	rctx.URLParams.Values[0] = hexutil.Encode(jobID)
	w, r = getHTTPReqAndResp(ctx)
	h = handler{}
	h.Job(w, r)
	assert.Equal(t, w.Code, http.StatusNotFound)
	assert.Contains(t, w.Body.String(), ErrJobNotFound.Error())

	// missing job
	did := testingidentity.GenerateRandomDID()
	ctx = context.WithValue(ctx, config.AccountHeaderKey, did.String())
	w, r = getHTTPReqAndResp(ctx)
	dispatcher := new(jobs.MockDispatcher)
	dispatcher.On("Job", did, jobID).Return(nil, errors.New("missing job")).Once()
	h = handler{srv: Service{dispatcher: dispatcher}}
	h.Job(w, r)
	assert.Equal(t, w.Code, http.StatusNotFound)
	assert.Contains(t, w.Body.String(), ErrJobNotFound.Error())

	// success
	w, r = getHTTPReqAndResp(ctx)
	dispatcher.On("Job", did, jobID).Return(new(gocelery.Job), nil).Once()
	h.Job(w, r)
	assert.Equal(t, w.Code, http.StatusOK)
	dispatcher.AssertExpectations(t)
}
