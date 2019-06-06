// +build unit

package coreapi

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/jobs"
	testingdocuments "github.com/centrifuge/go-centrifuge/testingutils/documents"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/go-chi/chi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestHandler_CreateDocument(t *testing.T) {
	data := map[string]interface{}{
		"scheme": "invoice",
		"data":   invoiceData(),
		"attributes": map[string]map[string]string{
			"string_test": {
				"type":  "invalid",
				"value": "hello, world",
			},
		},
	}

	d, err := json.Marshal(data)
	assert.NoError(t, err)
	r := httptest.NewRequest("POST", "/documents", bytes.NewReader(d))
	w := httptest.NewRecorder()

	h := handler{}
	h.CreateDocument(w, r)
	assert.Equal(t, w.Code, http.StatusBadRequest)
	assert.Contains(t, w.Body.String(), "not a valid attribute")

	data = map[string]interface{}{
		"scheme": "invoice",
		"data":   invoiceData(),
		"attributes": map[string]map[string]string{
			"string_test": {
				"type":  "string",
				"value": "hello, world",
			},
		},
	}
	d, err = json.Marshal(data)
	assert.NoError(t, err)
	docSrv := new(testingdocuments.MockService)
	srv := Service{docService: docSrv}
	h = handler{srv: srv}
	docSrv.On("CreateModel", mock.Anything, mock.Anything).Return(nil, jobs.NilJobID(), errors.New("failed to create model"))
	r = httptest.NewRequest("POST", "/documents", bytes.NewReader(d))
	w = httptest.NewRecorder()
	h.CreateDocument(w, r)
	assert.Equal(t, w.Code, http.StatusBadRequest)
	assert.Contains(t, w.Body.String(), "failed to create model")
	docSrv.AssertExpectations(t)

	m := new(testingdocuments.MockModel)
	m.On("GetData").Return(data)
	m.On("Scheme").Return("invoice")
	m.On("GetAttributes").Return(nil)
	m.On("GetCollaborators", mock.Anything).Return(documents.CollaboratorsAccess{}, errors.New("failed to get collaborators"))
	docSrv = new(testingdocuments.MockService)
	srv = Service{docService: docSrv}
	h = handler{srv: srv}
	docSrv.On("CreateModel", mock.Anything, mock.Anything).Return(m, jobs.NewJobID(), nil)
	r = httptest.NewRequest("POST", "/documents", bytes.NewReader(d))
	w = httptest.NewRecorder()
	h.CreateDocument(w, r)
	assert.Equal(t, w.Code, http.StatusInternalServerError)
	assert.Contains(t, w.Body.String(), "failed to get collaborators")
	m.AssertExpectations(t)
	docSrv.AssertExpectations(t)

	m = new(testingdocuments.MockModel)
	m.On("GetCollaborators", mock.Anything).Return(documents.CollaboratorsAccess{}, nil).Once()
	m.On("GetData").Return(data)
	m.On("Scheme").Return("invoice")
	m.On("ID").Return(utils.RandomSlice(32)).Once()
	m.On("CurrentVersion").Return(utils.RandomSlice(32)).Once()
	m.On("Author").Return(nil, errors.New("somerror"))
	m.On("Timestamp").Return(nil, errors.New("somerror"))
	m.On("NFTs").Return(nil)
	m.On("GetAttributes").Return(nil)
	docSrv = new(testingdocuments.MockService)
	srv = Service{docService: docSrv}
	h = handler{srv: srv}
	docSrv.On("CreateModel", mock.Anything, mock.Anything).Return(m, jobs.NewJobID(), nil)
	r = httptest.NewRequest("POST", "/documents", bytes.NewReader(d))
	w = httptest.NewRecorder()
	h.CreateDocument(w, r)
	assert.Equal(t, w.Code, http.StatusCreated)
	m.AssertExpectations(t)
	docSrv.AssertExpectations(t)
}

func TestHandler_UpdateDocument(t *testing.T) {
	id := hexutil.Encode(utils.RandomSlice(32))
	data := map[string]interface{}{
		"scheme":      "invoice",
		"document_id": id,
		"data":        invoiceData(),
		"attributes": map[string]map[string]string{
			"string_test": {
				"type":  "invalid",
				"value": "hello, world",
			},
		},
	}

	d, err := json.Marshal(data)
	assert.NoError(t, err)
	r := httptest.NewRequest("PUT", "/documents", bytes.NewReader(d))
	w := httptest.NewRecorder()

	h := handler{}
	h.UpdateDocument(w, r)
	assert.Equal(t, w.Code, http.StatusBadRequest)
	assert.Contains(t, w.Body.String(), "not a valid attribute")

	data = map[string]interface{}{
		"scheme":      "invoice",
		"data":        invoiceData(),
		"document_id": id,
		"attributes": map[string]map[string]string{
			"string_test": {
				"type":  "string",
				"value": "hello, world",
			},
		},
	}
	d, err = json.Marshal(data)
	assert.NoError(t, err)
	docSrv := new(testingdocuments.MockService)
	srv := Service{docService: docSrv}
	h = handler{srv: srv}
	docSrv.On("UpdateModel", mock.Anything, mock.Anything).Return(nil, jobs.NilJobID(), errors.New("failed to update model"))
	r = httptest.NewRequest("PUT", "/documents", bytes.NewReader(d))
	w = httptest.NewRecorder()
	h.UpdateDocument(w, r)
	assert.Equal(t, w.Code, http.StatusBadRequest)
	assert.Contains(t, w.Body.String(), "failed to update model")
	docSrv.AssertExpectations(t)

	m := new(testingdocuments.MockModel)
	m.On("GetData").Return(data)
	m.On("Scheme").Return("invoice")
	m.On("GetAttributes").Return(nil)
	m.On("GetCollaborators", mock.Anything).Return(documents.CollaboratorsAccess{}, errors.New("failed to get collaborators"))
	docSrv = new(testingdocuments.MockService)
	srv = Service{docService: docSrv}
	h = handler{srv: srv}
	docSrv.On("UpdateModel", mock.Anything, mock.Anything).Return(m, jobs.NewJobID(), nil)
	r = httptest.NewRequest("PUT", "/documents", bytes.NewReader(d))
	w = httptest.NewRecorder()
	h.UpdateDocument(w, r)
	assert.Equal(t, w.Code, http.StatusInternalServerError)
	assert.Contains(t, w.Body.String(), "failed to get collaborators")
	m.AssertExpectations(t)
	docSrv.AssertExpectations(t)

	m = new(testingdocuments.MockModel)
	m.On("GetCollaborators", mock.Anything).Return(documents.CollaboratorsAccess{}, nil).Once()
	m.On("GetData").Return(data)
	m.On("Scheme").Return("invoice")
	m.On("ID").Return(utils.RandomSlice(32)).Once()
	m.On("CurrentVersion").Return(utils.RandomSlice(32)).Once()
	m.On("Author").Return(nil, errors.New("somerror"))
	m.On("Timestamp").Return(nil, errors.New("somerror"))
	m.On("NFTs").Return(nil)
	m.On("GetAttributes").Return(nil)
	docSrv = new(testingdocuments.MockService)
	srv = Service{docService: docSrv}
	h = handler{srv: srv}
	docSrv.On("UpdateModel", mock.Anything, mock.Anything).Return(m, jobs.NewJobID(), nil)
	r = httptest.NewRequest("PUT", "/documents", bytes.NewReader(d))
	w = httptest.NewRecorder()
	h.UpdateDocument(w, r)
	assert.Equal(t, w.Code, http.StatusCreated)
	m.AssertExpectations(t)
	docSrv.AssertExpectations(t)
}

func TestHandler_GetDocument(t *testing.T) {
	getHTTPReqAndResp := func(ctx context.Context) (*httptest.ResponseRecorder, *http.Request) {
		return httptest.NewRecorder(), httptest.NewRequest("GET", "/documents/{document_id}", nil).WithContext(ctx)
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
		h.GetDocument(w, r)
		assert.Equal(t, w.Code, http.StatusBadRequest)
		assert.Contains(t, w.Body.String(), ErrInvalidDocumentID.Error())
	}

	// missing document
	id := utils.RandomSlice(32)
	rctx.URLParams.Values[0] = hexutil.Encode(id)
	docSrv := new(testingdocuments.MockService)
	docSrv.On("GetCurrentVersion", id).Return(nil, errors.New("failed"))
	h = handler{srv: Service{docService: docSrv}}
	w, r := getHTTPReqAndResp(ctx)
	h.GetDocument(w, r)
	assert.Equal(t, w.Code, http.StatusNotFound)
	assert.Contains(t, w.Body.String(), ErrDocumentNotFound.Error())
	docSrv.AssertExpectations(t)

	// failed doc response
	data := map[string]interface{}{
		"scheme":      "invoice",
		"data":        invoiceData(),
		"document_id": id,
		"attributes": map[string]map[string]string{
			"string_test": {
				"type":  "string",
				"value": "hello, world",
			},
		},
	}
	m := new(testingdocuments.MockModel)
	m.On("GetData").Return(data)
	m.On("Scheme").Return("invoice")
	m.On("GetAttributes").Return(nil)
	m.On("GetCollaborators", mock.Anything).Return(documents.CollaboratorsAccess{}, errors.New("failed to get collaborators"))
	docSrv = new(testingdocuments.MockService)
	docSrv.On("GetCurrentVersion", id).Return(m, nil)
	h = handler{srv: Service{docService: docSrv}}
	w, r = getHTTPReqAndResp(ctx)
	h.GetDocument(w, r)
	assert.Equal(t, w.Code, http.StatusInternalServerError)
	assert.Contains(t, w.Body.String(), "failed to get collaborators")
	docSrv.AssertExpectations(t)
	m.AssertExpectations(t)

	// success
	m = new(testingdocuments.MockModel)
	m.On("GetCollaborators", mock.Anything).Return(documents.CollaboratorsAccess{}, nil).Once()
	m.On("GetData").Return(data)
	m.On("Scheme").Return("invoice")
	m.On("ID").Return(utils.RandomSlice(32)).Once()
	m.On("CurrentVersion").Return(utils.RandomSlice(32)).Once()
	m.On("Author").Return(nil, errors.New("somerror"))
	m.On("Timestamp").Return(nil, errors.New("somerror"))
	m.On("NFTs").Return(nil)
	m.On("GetAttributes").Return(nil)
	docSrv = new(testingdocuments.MockService)
	docSrv.On("GetCurrentVersion", id).Return(m, nil)
	h = handler{srv: Service{docService: docSrv}}
	w, r = getHTTPReqAndResp(ctx)
	h.GetDocument(w, r)
	assert.Equal(t, w.Code, http.StatusOK)
	docSrv.AssertExpectations(t)
	m.AssertExpectations(t)
}

func TestHandler_GetDocumentVersion(t *testing.T) {
	getHTTPReqAndResp := func(ctx context.Context) (*httptest.ResponseRecorder, *http.Request) {
		return httptest.NewRecorder(), httptest.NewRequest("GET", "/documents/{document_id}/versions/{version_id}", nil).WithContext(ctx)
	}

	// empty document_id and invalid
	rctx := chi.NewRouteContext()
	rctx.URLParams.Keys = make([]string, 2, 2)
	rctx.URLParams.Values = make([]string, 2, 2)
	rctx.URLParams.Keys[0] = "document_id"
	rctx.URLParams.Keys[1] = "version_id"
	ctx := context.WithValue(context.Background(), chi.RouteCtxKey, rctx)
	h := handler{}

	for _, id := range []string{"", "invalid"} {
		rctx.URLParams.Values[0] = id
		rctx.URLParams.Values[1] = id
		w, r := getHTTPReqAndResp(ctx)
		h.GetDocumentVersion(w, r)
		assert.Equal(t, w.Code, http.StatusBadRequest)
		assert.Contains(t, w.Body.String(), ErrInvalidDocumentID.Error())
	}

	// missing document
	id := utils.RandomSlice(32)
	vid := utils.RandomSlice(32)
	rctx.URLParams.Values[0] = hexutil.Encode(id)
	rctx.URLParams.Values[1] = hexutil.Encode(vid)
	docSrv := new(testingdocuments.MockService)
	docSrv.On("GetVersion", id, vid).Return(nil, errors.New("failed"))
	h = handler{srv: Service{docService: docSrv}}
	w, r := getHTTPReqAndResp(ctx)
	h.GetDocumentVersion(w, r)
	assert.Equal(t, w.Code, http.StatusNotFound)
	assert.Contains(t, w.Body.String(), ErrDocumentNotFound.Error())
	docSrv.AssertExpectations(t)

	// failed doc response
	data := map[string]interface{}{
		"scheme":      "invoice",
		"data":        invoiceData(),
		"document_id": id,
		"attributes": map[string]map[string]string{
			"string_test": {
				"type":  "string",
				"value": "hello, world",
			},
		},
	}
	m := new(testingdocuments.MockModel)
	m.On("GetData").Return(data)
	m.On("Scheme").Return("invoice")
	m.On("GetAttributes").Return(nil)
	m.On("GetCollaborators", mock.Anything).Return(documents.CollaboratorsAccess{}, errors.New("failed to get collaborators"))
	docSrv = new(testingdocuments.MockService)
	docSrv.On("GetVersion", id, vid).Return(m, nil)
	h = handler{srv: Service{docService: docSrv}}
	w, r = getHTTPReqAndResp(ctx)
	h.GetDocumentVersion(w, r)
	assert.Equal(t, w.Code, http.StatusInternalServerError)
	assert.Contains(t, w.Body.String(), "failed to get collaborators")
	docSrv.AssertExpectations(t)
	m.AssertExpectations(t)

	// success
	m = new(testingdocuments.MockModel)
	m.On("GetCollaborators", mock.Anything).Return(documents.CollaboratorsAccess{}, nil).Once()
	m.On("GetData").Return(data)
	m.On("Scheme").Return("invoice")
	m.On("ID").Return(utils.RandomSlice(32)).Once()
	m.On("CurrentVersion").Return(utils.RandomSlice(32)).Once()
	m.On("Author").Return(nil, errors.New("somerror"))
	m.On("Timestamp").Return(nil, errors.New("somerror"))
	m.On("NFTs").Return(nil)
	m.On("GetAttributes").Return(nil)
	docSrv = new(testingdocuments.MockService)
	docSrv.On("GetVersion", id, vid).Return(m, nil)
	h = handler{srv: Service{docService: docSrv}}
	w, r = getHTTPReqAndResp(ctx)
	h.GetDocumentVersion(w, r)
	assert.Equal(t, w.Code, http.StatusOK)
	docSrv.AssertExpectations(t)
	m.AssertExpectations(t)
}
