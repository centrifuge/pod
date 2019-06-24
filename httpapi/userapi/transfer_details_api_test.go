// +build unit

package userapi

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/jobs"
	"github.com/centrifuge/go-centrifuge/testingutils/documents"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/go-chi/chi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestHandler_CreateTransferDetail(t *testing.T) {
	docSrv := new(testingdocuments.MockService)
	transferSrv := new(MockTransferService)
	srv := Service{docSrv: docSrv, transferDetailsService: transferSrv}
	h := handler{srv: srv}

	docID := hexutil.Encode(utils.RandomSlice(32))
	data := map[string]interface{}{
		"document_id": docID,
		"data":        transferData(),
	}
	d, err := json.Marshal(data)
	assert.NoError(t, err)

	getHTTPReqAndResp := func(ctx context.Context, b io.Reader) (*httptest.ResponseRecorder, *http.Request) {
		return httptest.NewRecorder(), httptest.NewRequest("POST", "/documents/{document_id}/transfer_details", b).WithContext(ctx)
	}
	rctx := chi.NewRouteContext()
	rctx.URLParams.Keys = make([]string, 1, 1)
	rctx.URLParams.Values = make([]string, 1, 1)
	rctx.URLParams.Keys[0] = "document_id"
	rctx.URLParams.Values[0] = docID
	ctx := context.WithValue(context.Background(), chi.RouteCtxKey, rctx)
	w, r := getHTTPReqAndResp(ctx, bytes.NewReader(d))

	// error in model
	transferSrv.On("CreateTransferDetail", mock.Anything, mock.Anything).Return(nil, jobs.NilJobID(), errors.New("failed to create model"))
	h.CreateTransferDetail(w, r)
	assert.Equal(t, w.Code, http.StatusBadRequest)
	assert.Contains(t, w.Body.String(), "failed to create model")

	// bad request
	m := new(testingdocuments.MockModel)
	m.On("GetCollaborators", mock.Anything).Return(documents.CollaboratorsAccess{}, nil).Once()
	m.On("GetData").Return(data)
	m.On("Scheme").Return("invoice")
	m.On("ID").Return(utils.RandomSlice(32)).Once()
	m.On("CurrentVersion").Return(utils.RandomSlice(32)).Once()
	m.On("Author").Return(nil, errors.New("somerror"))
	m.On("Timestamp").Return(nil, errors.New("somerror"))
	m.On("NFTs").Return(nil)
	m.On("GetAttributes").Return(nil)
	transferSrv.On("CreateTransferDetail", mock.Anything, mock.Anything).Return(m, jobs.NewJobID(), nil)
	h.CreateTransferDetail(w, r)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}
