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
	"github.com/centrifuge/go-centrifuge/documents/purchaseorder"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/httpapi/coreapi"
	"github.com/centrifuge/go-centrifuge/jobs"
	testingdocuments "github.com/centrifuge/go-centrifuge/testingutils/documents"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/go-chi/chi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func poData() map[string]interface{} {
	return map[string]interface{}{
		"number":         "12345",
		"status":         "unpaid",
		"total_amount":   "12.345",
		"recipient":      "0xBAEb33a61f05e6F269f1c4b4CFF91A901B54DaF7",
		"date_sent":      "2019-05-24T14:48:44.308854Z", // rfc3339nano
		"date_confirmed": "2019-05-24T14:48:44Z",        // rfc3339
		"currency":       "EUR",
		"attachments": []map[string]interface{}{
			{
				"name":      "test",
				"file_type": "pdf",
				"size":      1000202,
				"data":      "0xBAEb33a61f05e6F269f1c4b4CFF91A901B54DaF7",
				"checksum":  "0xBAEb33a61f05e6F269f1c4b4CFF91A901B54DaF3",
			},
		},
	}
}

func marshall(t *testing.T, d map[string]interface{}) []byte {
	data, err := json.Marshal(d)
	assert.NoError(t, err)
	return data
}

func TestHandler_CreatePurchaseOrder(t *testing.T) {
	data := map[string]interface{}{
		"data": poData(),
		"attributes": map[string]map[string]string{
			"string_test": {
				"type":  "invalid",
				"value": "hello, world",
			},
		},
	}

	getHTTPReqAndResp := func(ctx context.Context, b io.Reader) (*httptest.ResponseRecorder, *http.Request) {
		return httptest.NewRecorder(), httptest.NewRequest("POST", "/purchase_orders", b).WithContext(ctx)
	}

	// empty body
	rctx := chi.NewRouteContext()
	ctx := context.WithValue(context.Background(), chi.RouteCtxKey, rctx)
	h := handler{}
	w, r := getHTTPReqAndResp(ctx, nil)
	h.CreatePurchaseOrder(w, r)
	assert.Equal(t, w.Code, http.StatusBadRequest)
	assert.Contains(t, w.Body.String(), "unexpected end of JSON input")

	// invalid body
	w, r = getHTTPReqAndResp(ctx, bytes.NewReader(marshall(t, data)))
	h.CreatePurchaseOrder(w, r)
	assert.Equal(t, w.Code, http.StatusBadRequest)
	assert.Contains(t, w.Body.String(), documents.ErrNotValidAttrType.Error())

	// failed response conversion
	data["attributes"] = map[string]map[string]string{
		"string_test": {
			"type":  "string",
			"value": "hello, world",
		},
	}
	m := new(testingdocuments.MockModel)
	m.On("GetData").Return(purchaseorder.Data{})
	m.On("Scheme").Return(purchaseorder.Scheme)
	m.On("GetAttributes").Return(nil)
	m.On("GetCollaborators", mock.Anything).Return(documents.CollaboratorsAccess{}, errors.New("failed to get collaborators"))
	docSrv := new(testingdocuments.MockService)
	srv := Service{coreAPISrv: newCoreAPIService(docSrv)}
	h = handler{srv: srv}
	docSrv.On("CreateModel", mock.Anything, mock.Anything).Return(m, jobs.NewJobID(), nil)
	w, r = getHTTPReqAndResp(ctx, bytes.NewReader(marshall(t, data)))
	h.CreatePurchaseOrder(w, r)
	assert.Equal(t, w.Code, http.StatusInternalServerError)
	assert.Contains(t, w.Body.String(), "failed to get collaborators")
	m.AssertExpectations(t)
	docSrv.AssertExpectations(t)

	// success
	m = new(testingdocuments.MockModel)
	m.On("GetCollaborators", mock.Anything).Return(documents.CollaboratorsAccess{}, nil).Once()
	m.On("GetData").Return(purchaseorder.Data{})
	m.On("Scheme").Return(purchaseorder.Scheme)
	m.On("ID").Return(utils.RandomSlice(32)).Once()
	m.On("CurrentVersion").Return(utils.RandomSlice(32)).Once()
	m.On("Author").Return(nil, errors.New("somerror"))
	m.On("Timestamp").Return(nil, errors.New("somerror"))
	m.On("NFTs").Return(nil)
	m.On("GetAttributes").Return(nil)
	docSrv = new(testingdocuments.MockService)
	srv = Service{coreAPISrv: newCoreAPIService(docSrv)}
	h = handler{srv: srv}
	docSrv.On("CreateModel", mock.Anything, mock.Anything).Return(m, jobs.NewJobID(), nil)
	w, r = getHTTPReqAndResp(ctx, bytes.NewReader(marshall(t, data)))
	h.CreatePurchaseOrder(w, r)
	assert.Equal(t, w.Code, http.StatusCreated)
	m.AssertExpectations(t)
	docSrv.AssertExpectations(t)
}

func TestHandler_GetPurchaseOrder(t *testing.T) {
	getHTTPReqAndResp := func(ctx context.Context) (*httptest.ResponseRecorder, *http.Request) {
		return httptest.NewRecorder(), httptest.NewRequest("GET", "/purchase_orders/{document_id}", nil).WithContext(ctx)
	}

	// empty document_id and invalid
	rctx := chi.NewRouteContext()
	rctx.URLParams.Keys = make([]string, 1, 1)
	rctx.URLParams.Values = make([]string, 1, 1)
	rctx.URLParams.Keys[0] = "document_id"
	ctx := context.WithValue(context.Background(), chi.RouteCtxKey, rctx)
	h := handler{}

	for _, id := range []string{"", "invalid"} {
		rctx.URLParams.Values[0] = id
		w, r := getHTTPReqAndResp(ctx)
		h.GetPurchaseOrder(w, r)
		assert.Equal(t, w.Code, http.StatusBadRequest)
		assert.Contains(t, w.Body.String(), coreapi.ErrInvalidDocumentID.Error())
	}

	// missing document
	id := utils.RandomSlice(32)
	rctx.URLParams.Values[0] = hexutil.Encode(id)
	docSrv := new(testingdocuments.MockService)
	docSrv.On("GetCurrentVersion", id).Return(nil, errors.New("failed"))
	h = handler{srv: Service{coreAPISrv: newCoreAPIService(docSrv)}}
	w, r := getHTTPReqAndResp(ctx)
	h.GetPurchaseOrder(w, r)
	assert.Equal(t, w.Code, http.StatusNotFound)
	assert.Contains(t, w.Body.String(), coreapi.ErrDocumentNotFound.Error())
	docSrv.AssertExpectations(t)

	// failed doc response
	m := new(testingdocuments.MockModel)
	m.On("GetData").Return(purchaseorder.Data{})
	m.On("Scheme").Return(purchaseorder.Scheme)
	m.On("GetAttributes").Return(nil)
	m.On("GetCollaborators", mock.Anything).Return(documents.CollaboratorsAccess{}, errors.New("failed to get collaborators"))
	docSrv = new(testingdocuments.MockService)
	docSrv.On("GetCurrentVersion", id).Return(m, nil)
	h = handler{srv: Service{coreAPISrv: newCoreAPIService(docSrv)}}
	w, r = getHTTPReqAndResp(ctx)
	h.GetPurchaseOrder(w, r)
	assert.Equal(t, w.Code, http.StatusInternalServerError)
	assert.Contains(t, w.Body.String(), "failed to get collaborators")
	docSrv.AssertExpectations(t)
	m.AssertExpectations(t)

	// success
	m = new(testingdocuments.MockModel)
	m.On("GetCollaborators", mock.Anything).Return(documents.CollaboratorsAccess{}, nil).Once()
	m.On("GetData").Return(purchaseorder.Data{})
	m.On("Scheme").Return(purchaseorder.Scheme)
	m.On("ID").Return(utils.RandomSlice(32)).Once()
	m.On("CurrentVersion").Return(utils.RandomSlice(32)).Once()
	m.On("Author").Return(nil, errors.New("somerror"))
	m.On("Timestamp").Return(nil, errors.New("somerror"))
	m.On("NFTs").Return(nil)
	m.On("GetAttributes").Return(nil)
	docSrv = new(testingdocuments.MockService)
	docSrv.On("GetCurrentVersion", id).Return(m, nil)
	h = handler{srv: Service{coreAPISrv: newCoreAPIService(docSrv)}}
	w, r = getHTTPReqAndResp(ctx)
	h.GetPurchaseOrder(w, r)
	assert.Equal(t, w.Code, http.StatusOK)
	docSrv.AssertExpectations(t)
	m.AssertExpectations(t)
}
