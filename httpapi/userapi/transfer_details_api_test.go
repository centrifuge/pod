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

	"github.com/centrifuge/go-centrifuge/extensions/transferdetails"
	"github.com/centrifuge/go-centrifuge/testingutils/identity"

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
	srv := Service{coreAPISrv: newCoreAPIService(docSrv), transferDetailsService: transferSrv}
	h := handler{srv: srv}

	docID := hexutil.Encode(utils.RandomSlice(32))

	data := map[string]interface{}{
		"status":         "open",
		"currency":       "EUR",
		"amount":         "300",
		"scheduled_date": "2018-09-26T23:12:37Z",
		"sender_id":      testingidentity.GenerateRandomDID().String(),
		"recipient_id":   testingidentity.GenerateRandomDID().String(),
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

	// success
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
	m.On("CalculateTransitionRulesFingerprint").Return(utils.RandomSlice(32), nil)

	w, r = getHTTPReqAndResp(ctx, bytes.NewReader(d))
	transferSrv = new(MockTransferService)
	srv = Service{coreAPISrv: newCoreAPIService(docSrv), transferDetailsService: transferSrv}
	h = handler{srv: srv}
	transferSrv.On("CreateTransferDetail", mock.Anything, mock.Anything).Return(m, jobs.NewJobID(), nil)
	h.CreateTransferDetail(w, r)
	assert.Equal(t, http.StatusAccepted, w.Code)
}

func TestHandler_GetTransferDetail(t *testing.T) {
	docSrv := new(testingdocuments.MockService)
	transferSrv := new(MockTransferService)
	srv := Service{coreAPISrv: newCoreAPIService(docSrv), transferDetailsService: transferSrv}
	h := handler{srv: srv}

	data := map[string]interface{}{
		"status":         "open",
		"currency":       "EUR",
		"amount":         "300",
		"scheduled_date": "2018-09-26T23:12:37Z",
		"sender_id":      testingidentity.GenerateRandomDID().String(),
		"recipient_id":   testingidentity.GenerateRandomDID().String(),
	}
	_, err := json.Marshal(data)
	assert.NoError(t, err)

	getHTTPReqAndResp := func(ctx context.Context) (*httptest.ResponseRecorder, *http.Request) {
		return httptest.NewRecorder(), httptest.NewRequest("GET", "/documents/{document_id}/transfer_details/{transfer_id}", nil).WithContext(ctx)
	}

	// empty hex string docID
	docID := ""
	transferID := hexutil.Encode(utils.RandomSlice(32))
	rctx := chi.NewRouteContext()
	rctx.URLParams.Keys = make([]string, 2, 2)
	rctx.URLParams.Values = make([]string, 2, 2)
	rctx.URLParams.Keys[0] = "document_id"
	rctx.URLParams.Values[0] = docID
	rctx.URLParams.Keys[1] = "transfer_id"
	rctx.URLParams.Values[1] = transferID
	ctx := context.WithValue(context.Background(), chi.RouteCtxKey, rctx)
	w, r := getHTTPReqAndResp(ctx)
	transferSrv.On("GetTransferDetail", mock.Anything, mock.Anything).Return(nil, jobs.NilJobID(), errors.New("failed to create model"))

	h.GetTransferDetail(w, r)
	assert.Equal(t, w.Code, http.StatusBadRequest)
	assert.Contains(t, w.Body.String(), "invalid document identifier")

	// missing document
	id := utils.RandomSlice(32)
	rctx.URLParams.Values[0] = hexutil.Encode(id)
	docSrv = new(testingdocuments.MockService)
	docSrv.On("GetCurrentVersion", id).Return(nil, errors.New("document not found"))
	srv = Service{coreAPISrv: newCoreAPIService(docSrv), transferDetailsService: transferSrv}
	h = handler{srv: srv}
	ctx = context.WithValue(context.Background(), chi.RouteCtxKey, rctx)
	w, r = getHTTPReqAndResp(ctx)

	h.GetTransferDetail(w, r)
	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "document not found")

	// success
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
	m.On("CalculateTransitionRulesFingerprint").Return(utils.RandomSlice(32), nil)

	w, r = getHTTPReqAndResp(ctx)
	d := &transferdetails.TransferDetail{
		Data: transferdetails.Data{
			TransferID: transferID,
		},
	}
	docSrv = new(testingdocuments.MockService)
	docSrv.On("GetCurrentVersion", id).Return(m, nil)
	transferSrv = new(MockTransferService)
	srv = Service{coreAPISrv: newCoreAPIService(docSrv), transferDetailsService: transferSrv}
	h = handler{srv: srv}
	transferSrv.On("DeriveTransferDetail", mock.Anything, m, mock.Anything).Return(d, m, nil)
	h.GetTransferDetail(w, r)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestHandler_UpdateTransferDetail(t *testing.T) {
	docSrv := new(testingdocuments.MockService)
	transferSrv := new(MockTransferService)
	srv := Service{coreAPISrv: newCoreAPIService(docSrv), transferDetailsService: transferSrv}
	h := handler{srv: srv}

	docID := hexutil.Encode(utils.RandomSlice(32))

	data := map[string]interface{}{
		"status":         "settled",
		"currency":       "EUR",
		"amount":         "300",
		"scheduled_date": "2018-09-26T23:12:37Z",
		"sender_id":      testingidentity.GenerateRandomDID().String(),
		"recipient_id":   testingidentity.GenerateRandomDID().String(),
	}
	d, err := json.Marshal(data)
	assert.NoError(t, err)

	getHTTPReqAndResp := func(ctx context.Context, b io.Reader) (*httptest.ResponseRecorder, *http.Request) {
		return httptest.NewRecorder(), httptest.NewRequest("POST", "/documents/{document_id}/transfer_detail/{transfer_id}", b).WithContext(ctx)
	}
	rctx := chi.NewRouteContext()
	rctx.URLParams.Keys = make([]string, 1, 1)
	rctx.URLParams.Values = make([]string, 1, 1)
	rctx.URLParams.Keys[0] = "document_id"
	rctx.URLParams.Values[0] = docID
	ctx := context.WithValue(context.Background(), chi.RouteCtxKey, rctx)
	w, r := getHTTPReqAndResp(ctx, bytes.NewReader(d))

	// error in model
	transferSrv.On("UpdateTransferDetail", mock.Anything, mock.Anything).Return(nil, jobs.NilJobID(), errors.New("failed to create model"))
	h.UpdateTransferDetail(w, r)
	assert.Equal(t, w.Code, http.StatusBadRequest)
	assert.Contains(t, w.Body.String(), "failed to create model")

	// success
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
	m.On("CalculateTransitionRulesFingerprint").Return(utils.RandomSlice(32), nil)

	w, r = getHTTPReqAndResp(ctx, bytes.NewReader(d))
	transferSrv = new(MockTransferService)
	srv = Service{coreAPISrv: newCoreAPIService(docSrv), transferDetailsService: transferSrv}
	h = handler{srv: srv}
	transferSrv.On("UpdateTransferDetail", mock.Anything, mock.Anything).Return(m, jobs.NewJobID(), nil)
	h.UpdateTransferDetail(w, r)
	assert.Equal(t, http.StatusAccepted, w.Code)
}
