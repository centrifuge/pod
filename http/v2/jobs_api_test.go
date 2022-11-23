//go:build unit

package v2

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/contextutil"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/jobs"
	testingcommons "github.com/centrifuge/go-centrifuge/testingutils/common"
	genericUtils "github.com/centrifuge/go-centrifuge/testingutils/generic"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/centrifuge/gocelery/v2"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/go-chi/chi"
	"github.com/stretchr/testify/assert"
)

func TestHandler_Job(t *testing.T) {
	service, mocks := getServiceWithMocks(t)
	ctx := context.Background()

	serviceContext := map[string]any{
		BootstrappedService: service,
	}

	router := chi.NewRouter()

	accountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	accountMock := config.NewAccountMock(t)
	accountMock.On("GetIdentity").
		Return(accountID)

	// Mimic the auth handler by adding the account to context.
	router.Use(func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			ctx = contextutil.WithAccount(request.Context(), accountMock)

			request = request.WithContext(ctx)

			h.ServeHTTP(writer, request)
		})
	})

	Register(serviceContext, router)

	testServer := httptest.NewServer(router)
	defer testServer.Close()

	jobID := utils.RandomSlice(32)

	testURL := fmt.Sprintf("%s/jobs/%s", testServer.URL, hexutil.Encode(jobID))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, testURL, nil)
	assert.NoError(t, err)

	job := &Job{
		ID:     jobID,
		Desc:   "job-desc",
		Runner: "job-runner",
		Overrides: map[string]any{
			"override1": "string-override",
		},
		Tasks: []*gocelery.Task{
			{
				RunnerFunc: "runner-func",
				Args: []any{
					"arg1",
					2,
				},
				Result: "result",
				Error:  "",
				Tries:  0,
				Delay:  time.Now(),
			},
		},
		ValidUntil: time.Now(),
		FinishedAt: time.Now(),
		Finished:   false,
	}

	genericUtils.GetMock[*jobs.DispatcherMock](mocks).On("Job", accountID, gocelery.JobID(jobID)).
		Return(job, nil).Once()

	res, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode)

	resBody, err := ioutil.ReadAll(res.Body)
	assert.NoError(t, err)

	buf := &bytes.Buffer{}
	enc := json.NewEncoder(buf)
	enc.SetEscapeHTML(true)

	err = enc.Encode(job)
	assert.NoError(t, err)

	jsonJob := buf.Bytes()

	assert.Equal(t, string(jsonJob), string(resBody))
}

func TestHandler_Job_InvalidJobIDParam(t *testing.T) {
	service, _ := getServiceWithMocks(t)
	ctx := context.Background()

	serviceContext := map[string]any{
		BootstrappedService: service,
	}

	router := chi.NewRouter()

	Register(serviceContext, router)

	testServer := httptest.NewServer(router)
	defer testServer.Close()

	jobID := "invalid-job-id-param"

	testURL := fmt.Sprintf("%s/jobs/%s", testServer.URL, jobID)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, testURL, nil)
	assert.NoError(t, err)

	res, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
}

func TestHandler_Job_NoAccount(t *testing.T) {
	service, _ := getServiceWithMocks(t)
	ctx := context.Background()

	serviceContext := map[string]any{
		BootstrappedService: service,
	}

	router := chi.NewRouter()

	Register(serviceContext, router)

	testServer := httptest.NewServer(router)
	defer testServer.Close()

	jobID := utils.RandomSlice(32)

	testURL := fmt.Sprintf("%s/jobs/%s", testServer.URL, hexutil.Encode(jobID))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, testURL, nil)
	assert.NoError(t, err)

	res, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, res.StatusCode)
}

func TestHandler_Job_DispatcherError(t *testing.T) {
	service, mocks := getServiceWithMocks(t)
	ctx := context.Background()

	serviceContext := map[string]any{
		BootstrappedService: service,
	}

	router := chi.NewRouter()

	accountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	accountMock := config.NewAccountMock(t)
	accountMock.On("GetIdentity").
		Return(accountID)

	// Mimic the auth handler by adding the account to context.
	router.Use(func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			ctx = contextutil.WithAccount(request.Context(), accountMock)

			request = request.WithContext(ctx)

			h.ServeHTTP(writer, request)
		})
	})

	Register(serviceContext, router)

	testServer := httptest.NewServer(router)
	defer testServer.Close()

	jobID := utils.RandomSlice(32)

	testURL := fmt.Sprintf("%s/jobs/%s", testServer.URL, hexutil.Encode(jobID))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, testURL, nil)
	assert.NoError(t, err)

	genericUtils.GetMock[*jobs.DispatcherMock](mocks).On("Job", accountID, gocelery.JobID(jobID)).
		Return(nil, errors.New("error")).Once()

	res, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, res.StatusCode)
}
