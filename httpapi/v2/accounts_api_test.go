// +build unit

package v2

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/centrifuge/go-centrifuge/config/configstore"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/go-chi/chi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestHandler_GenerateAccount(t *testing.T) {
	getHTTPReqAndResp := func(ctx context.Context, body io.Reader) (*httptest.ResponseRecorder, *http.Request) {
		return httptest.NewRecorder(), httptest.NewRequest("POST", "/accounts/generate", body).WithContext(ctx)
	}

	// empty body
	rctx := chi.NewRouteContext()
	ctx := context.WithValue(context.Background(), chi.RouteCtxKey, rctx)
	h := handler{}
	w, r := getHTTPReqAndResp(ctx, nil)
	h.GenerateAccount(w, r)
	assert.Equal(t, w.Code, http.StatusBadRequest)
	assert.Contains(t, w.Body.String(), "unexpected end of JSON input")

	// failed generation
	data := map[string]interface{}{
		"centrifuge_chain_account": map[string]string{
			"id":            "0xc81ebbec0559a6acf184535eb19da51ed3ed8c4ac65323999482aaf9b6696e27",
			"secret":        "0xc166b100911b1e9f780bb66d13badf2c1edbe94a1220f1a0584c09490158be31",
			"ss_58_address": "5Gb6Zfe8K8NSKrkFLCgqs8LUdk7wKweXM5pN296jVqDpdziR",
		},
	}
	d, err := json.Marshal(data)
	assert.NoError(t, err)
	srv := new(configstore.MockService)
	srv.On("GenerateAccountAsync", mock.Anything).Return(nil, nil, errors.New("failed to generate account")).Once()
	h.srv.accountSrv = srv
	w, r = getHTTPReqAndResp(ctx, bytes.NewReader(d))
	h.GenerateAccount(w, r)
	assert.Equal(t, w.Code, http.StatusInternalServerError)
	assert.Contains(t, w.Body.String(), "failed to generate account")
	srv.AssertExpectations(t)

	// success
	did := utils.RandomSlice(20)
	jobID := utils.RandomSlice(32)
	srv.On("GenerateAccountAsync", mock.Anything).Return(did, jobID, nil).Once()
	h.srv.accountSrv = srv
	w, r = getHTTPReqAndResp(ctx, bytes.NewReader(d))
	h.GenerateAccount(w, r)
	srv.AssertExpectations(t)
	assert.Equal(t, w.Code, http.StatusOK)
	assert.Contains(t, w.Body.String(), hexutil.Encode(did))
	assert.Contains(t, w.Body.String(), hexutil.Encode(jobID))
}
