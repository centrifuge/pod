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

	"github.com/centrifuge/go-centrifuge/config/configstore"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/documents/invoice"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/httpapi/coreapi"
	"github.com/centrifuge/go-centrifuge/jobs"
	"github.com/centrifuge/go-centrifuge/nft"
	"github.com/centrifuge/go-centrifuge/testingutils/config"
	"github.com/centrifuge/go-centrifuge/testingutils/documents"
	"github.com/centrifuge/go-centrifuge/testingutils/nfts"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/go-chi/chi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func invData() map[string]interface{} {
	return map[string]interface{}{
		"number":         "12345",
		"status":         "unpaid",
		"recipient":      "0xBAEb33a61f05e6F269f1c4b4CFF91A901B54DaF7",
		"date_confirmed": "2019-05-24T14:48:44Z", // rfc3339
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

func marshall(t *testing.T, data interface{}) []byte {
	d, err := json.Marshal(data)
	assert.NoError(t, err)
	return d
}

func TestHandler_CreateInvoice(t *testing.T) {
	data := map[string]interface{}{
		"data": invData(),
		"attributes": map[string]map[string]string{
			"string_test": {
				"type":  "invalid",
				"value": "hello, world",
			},
		},
	}

	getHTTPReqAndResp := func(ctx context.Context, b io.Reader) (*httptest.ResponseRecorder, *http.Request) {
		return httptest.NewRecorder(), httptest.NewRequest("POST", "/invoices", b).WithContext(ctx)
	}

	// empty body
	rctx := chi.NewRouteContext()
	ctx := context.WithValue(context.Background(), chi.RouteCtxKey, rctx)
	h := handler{}
	w, r := getHTTPReqAndResp(ctx, nil)
	h.CreateInvoice(w, r)
	assert.Equal(t, w.Code, http.StatusBadRequest)
	assert.Contains(t, w.Body.String(), "unexpected end of JSON input")

	// invalid body
	w, r = getHTTPReqAndResp(ctx, bytes.NewReader(marshall(t, data)))
	h.CreateInvoice(w, r)
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
	m.On("GetData").Return(invoice.Data{})
	m.On("Scheme").Return(invoice.Scheme)
	m.On("GetAttributes").Return(nil)
	m.On("GetCollaborators", mock.Anything).Return(documents.CollaboratorsAccess{}, errors.New("failed to get collaborators"))
	docSrv := new(testingdocuments.MockService)
	srv := Service{coreAPISrv: newCoreAPIService(docSrv)}
	h = handler{srv: srv}
	docSrv.On("CreateModel", mock.Anything, mock.Anything).Return(m, jobs.NewJobID(), nil)
	w, r = getHTTPReqAndResp(ctx, bytes.NewReader(marshall(t, data)))
	h.CreateInvoice(w, r)
	assert.Equal(t, w.Code, http.StatusInternalServerError)
	assert.Contains(t, w.Body.String(), "failed to get collaborators")
	m.AssertExpectations(t)
	docSrv.AssertExpectations(t)

	// success
	m = new(testingdocuments.MockModel)
	m.On("GetCollaborators", mock.Anything).Return(documents.CollaboratorsAccess{}, nil).Once()
	m.On("GetData").Return(invoice.Data{})
	m.On("Scheme").Return(invoice.Scheme)
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
	h.CreateInvoice(w, r)
	assert.Equal(t, w.Code, http.StatusAccepted)
	m.AssertExpectations(t)
	docSrv.AssertExpectations(t)
}

func TestHandler_GetInvoice(t *testing.T) {
	getHTTPReqAndResp := func(ctx context.Context) (*httptest.ResponseRecorder, *http.Request) {
		return httptest.NewRecorder(), httptest.NewRequest("GET", "/invoices/{document_id}", nil).WithContext(ctx)
	}

	rctx := chi.NewRouteContext()
	rctx.URLParams.Keys = make([]string, 1, 1)
	rctx.URLParams.Values = make([]string, 1, 1)
	rctx.URLParams.Keys[0] = "document_id"
	ctx := context.WithValue(context.Background(), chi.RouteCtxKey, rctx)
	h := handler{}

	// empty document_id and invalid
	for _, id := range []string{"", "invalid"} {
		rctx.URLParams.Values[0] = id
		w, r := getHTTPReqAndResp(ctx)
		h.GetInvoice(w, r)
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
	h.GetInvoice(w, r)
	assert.Equal(t, w.Code, http.StatusNotFound)
	assert.Contains(t, w.Body.String(), coreapi.ErrDocumentNotFound.Error())
	docSrv.AssertExpectations(t)

	// failed doc response
	m := new(testingdocuments.MockModel)
	m.On("GetData").Return(invoice.Data{})
	m.On("Scheme").Return(invoice.Scheme)
	m.On("GetAttributes").Return(nil)
	m.On("GetCollaborators", mock.Anything).Return(documents.CollaboratorsAccess{}, errors.New("failed to get collaborators"))
	docSrv = new(testingdocuments.MockService)
	docSrv.On("GetCurrentVersion", id).Return(m, nil)
	h = handler{srv: Service{coreAPISrv: newCoreAPIService(docSrv)}}
	w, r = getHTTPReqAndResp(ctx)
	h.GetInvoice(w, r)
	assert.Equal(t, w.Code, http.StatusInternalServerError)
	assert.Contains(t, w.Body.String(), "failed to get collaborators")
	docSrv.AssertExpectations(t)
	m.AssertExpectations(t)

	// success
	m = new(testingdocuments.MockModel)
	m.On("GetCollaborators", mock.Anything).Return(documents.CollaboratorsAccess{}, nil).Once()
	m.On("GetData").Return(invoice.Data{})
	m.On("Scheme").Return(invoice.Scheme)
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
	h.GetInvoice(w, r)
	assert.Equal(t, w.Code, http.StatusOK)
	docSrv.AssertExpectations(t)
	m.AssertExpectations(t)
}

func TestHandler_UpdateInvoice(t *testing.T) {
	getHTTPReqAndResp := func(ctx context.Context, b io.Reader) (*httptest.ResponseRecorder, *http.Request) {
		return httptest.NewRecorder(), httptest.NewRequest("PUT", "/invoice/{document_id}", b).WithContext(ctx)
	}
	// empty document_id and invalid id
	rctx := chi.NewRouteContext()
	rctx.URLParams.Keys = make([]string, 1, 1)
	rctx.URLParams.Values = make([]string, 1, 1)
	rctx.URLParams.Keys[0] = "document_id"
	ctx := context.WithValue(context.Background(), chi.RouteCtxKey, rctx)
	h := handler{}
	for _, id := range []string{"", "invalid"} {
		rctx.URLParams.Values[0] = id
		w, r := getHTTPReqAndResp(ctx, nil)
		h.UpdateInvoice(w, r)
		assert.Equal(t, w.Code, http.StatusBadRequest)
		assert.Contains(t, w.Body.String(), coreapi.ErrInvalidDocumentID.Error())
	}

	// empty body
	id := hexutil.Encode(utils.RandomSlice(32))
	rctx.URLParams.Values[0] = id
	w, r := getHTTPReqAndResp(ctx, nil)
	h.UpdateInvoice(w, r)
	assert.Equal(t, w.Code, http.StatusBadRequest)
	assert.Contains(t, w.Body.String(), "unexpected end of JSON input")

	// failed conversion
	data := map[string]interface{}{
		"data": invData(),
		"attributes": map[string]map[string]string{
			"string_test": {
				"type":  "invalid",
				"value": "hello, world",
			},
		},
	}

	d, err := json.Marshal(data)
	assert.NoError(t, err)
	w, r = getHTTPReqAndResp(ctx, bytes.NewReader(d))
	h.UpdateInvoice(w, r)
	assert.Equal(t, w.Code, http.StatusBadRequest)
	assert.Contains(t, w.Body.String(), documents.ErrNotValidAttrType.Error())

	// failed to update
	data["attributes"] = map[string]map[string]string{
		"string_test": {
			"type":  "string",
			"value": "hello, world",
		},
	}
	d, err = json.Marshal(data)
	assert.NoError(t, err)
	docSrv := new(testingdocuments.MockService)
	srv := Service{coreAPISrv: newCoreAPIService(docSrv)}
	h = handler{srv: srv}
	docSrv.On("UpdateModel", mock.Anything, mock.Anything).Return(nil, jobs.NilJobID(), errors.New("failed to update model"))
	w, r = getHTTPReqAndResp(ctx, bytes.NewReader(d))
	h.UpdateInvoice(w, r)
	assert.Equal(t, w.Code, http.StatusBadRequest)
	assert.Contains(t, w.Body.String(), "failed to update model")
	docSrv.AssertExpectations(t)

	// failed response conversion
	m := new(testingdocuments.MockModel)
	m.On("GetData").Return(invoice.Data{})
	m.On("Scheme").Return(invoice.Scheme)
	m.On("GetAttributes").Return(nil)
	m.On("GetCollaborators", mock.Anything).Return(documents.CollaboratorsAccess{}, errors.New("failed to get collaborators"))
	docSrv = new(testingdocuments.MockService)
	srv = Service{coreAPISrv: newCoreAPIService(docSrv)}
	h = handler{srv: srv}
	docSrv.On("UpdateModel", mock.Anything, mock.Anything).Return(m, jobs.NewJobID(), nil)
	w, r = getHTTPReqAndResp(ctx, bytes.NewReader(d))
	h.UpdateInvoice(w, r)
	assert.Equal(t, w.Code, http.StatusInternalServerError)
	assert.Contains(t, w.Body.String(), "failed to get collaborators")
	docSrv.AssertExpectations(t)
	m.AssertExpectations(t)

	// success
	m = new(testingdocuments.MockModel)
	m.On("GetCollaborators", mock.Anything).Return(documents.CollaboratorsAccess{}, nil).Once()
	m.On("GetData").Return(invoice.Data{})
	m.On("Scheme").Return(invoice.Scheme)
	m.On("ID").Return(utils.RandomSlice(32)).Once()
	m.On("CurrentVersion").Return(utils.RandomSlice(32)).Once()
	m.On("Author").Return(nil, errors.New("somerror"))
	m.On("Timestamp").Return(nil, errors.New("somerror"))
	m.On("NFTs").Return(nil)
	m.On("GetAttributes").Return(nil)
	docSrv = new(testingdocuments.MockService)
	srv = Service{coreAPISrv: newCoreAPIService(docSrv)}
	h = handler{srv: srv}
	docSrv.On("UpdateModel", mock.Anything, mock.Anything).Return(m, jobs.NewJobID(), nil)
	w, r = getHTTPReqAndResp(ctx, bytes.NewReader(d))
	h.UpdateInvoice(w, r)
	assert.Equal(t, w.Code, http.StatusAccepted)
	m.AssertExpectations(t)
	docSrv.AssertExpectations(t)
}

func TestHandler_GetInvoiceVersion(t *testing.T) {
	getHTTPReqAndResp := func(ctx context.Context) (*httptest.ResponseRecorder, *http.Request) {
		return httptest.NewRecorder(), httptest.NewRequest("GET", "/invoices/{document_id}/versions/{version_id}", nil).WithContext(ctx)
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
		h.GetInvoiceVersion(w, r)
		assert.Equal(t, w.Code, http.StatusBadRequest)
		assert.Contains(t, w.Body.String(), coreapi.ErrInvalidDocumentID.Error())
	}

	// missing document
	id := utils.RandomSlice(32)
	vid := utils.RandomSlice(32)
	rctx.URLParams.Values[0] = hexutil.Encode(id)
	rctx.URLParams.Values[1] = hexutil.Encode(vid)
	docSrv := new(testingdocuments.MockService)
	docSrv.On("GetVersion", id, vid).Return(nil, errors.New("failed"))
	h = handler{srv: Service{coreAPISrv: newCoreAPIService(docSrv)}}
	w, r := getHTTPReqAndResp(ctx)
	h.GetInvoiceVersion(w, r)
	assert.Equal(t, w.Code, http.StatusNotFound)
	assert.Contains(t, w.Body.String(), coreapi.ErrDocumentNotFound.Error())
	docSrv.AssertExpectations(t)

	// failed doc response
	m := new(testingdocuments.MockModel)
	m.On("GetData").Return(invoice.Data{})
	m.On("Scheme").Return(invoice.Scheme)
	m.On("GetAttributes").Return(nil)
	m.On("GetCollaborators", mock.Anything).Return(documents.CollaboratorsAccess{}, errors.New("failed to get collaborators"))
	docSrv = new(testingdocuments.MockService)
	docSrv.On("GetVersion", id, vid).Return(m, nil)
	h = handler{srv: Service{coreAPISrv: newCoreAPIService(docSrv)}}
	w, r = getHTTPReqAndResp(ctx)
	h.GetInvoiceVersion(w, r)
	assert.Equal(t, w.Code, http.StatusInternalServerError)
	assert.Contains(t, w.Body.String(), "failed to get collaborators")
	docSrv.AssertExpectations(t)
	m.AssertExpectations(t)

	// success
	m = new(testingdocuments.MockModel)
	m.On("GetCollaborators", mock.Anything).Return(documents.CollaboratorsAccess{}, nil).Once()
	m.On("GetData").Return(invoice.Data{})
	m.On("Scheme").Return(invoice.Scheme)
	m.On("ID").Return(utils.RandomSlice(32)).Once()
	m.On("CurrentVersion").Return(utils.RandomSlice(32)).Once()
	m.On("Author").Return(nil, errors.New("somerror"))
	m.On("Timestamp").Return(nil, errors.New("somerror"))
	m.On("NFTs").Return(nil)
	m.On("GetAttributes").Return(nil)
	docSrv = new(testingdocuments.MockService)
	docSrv.On("GetVersion", id, vid).Return(m, nil)
	h = handler{srv: Service{coreAPISrv: newCoreAPIService(docSrv)}}
	w, r = getHTTPReqAndResp(ctx)
	h.GetInvoiceVersion(w, r)
	assert.Equal(t, w.Code, http.StatusOK)
	docSrv.AssertExpectations(t)
	m.AssertExpectations(t)
}

func TestHandler_MintInvoiceUnpaidNFT(t *testing.T) {
	getHTTPReqAndResp := func(ctx context.Context, b io.Reader) (*httptest.ResponseRecorder, *http.Request) {
		return httptest.NewRecorder(), httptest.NewRequest("POST", "/v1/invoices/{document_id}/mint/unpaid", b).WithContext(ctx)
	}

	srv := new(testingnfts.MockNFTService)
	mc := configstore.MockService{}
	mc.On("GetConfig").Return(cfg, nil)
	h := handler{srv: Service{coreAPISrv: coreapi.NewService(nil, nil, srv, nil), config: mc}}
	rctx := chi.NewRouteContext()
	rctx.URLParams.Keys = make([]string, 1, 1)
	rctx.URLParams.Values = make([]string, 1, 1)
	rctx.URLParams.Keys[0] = "document_id"
	c := testingconfig.CreateAccountContext(t, cfg)
	ctx := context.WithValue(c, chi.RouteCtxKey, rctx)

	// invalid docID
	w, r := getHTTPReqAndResp(ctx, nil)
	h.MintInvoiceUnpaidNFT(w, r)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), coreapi.ErrInvalidDocumentID.Error())

	// empty data
	docID := utils.RandomSlice(32)
	rctx.URLParams.Values[0] = hexutil.Encode(docID)
	w, r = getHTTPReqAndResp(ctx, nil)
	h.MintInvoiceUnpaidNFT(w, r)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "unexpected end of JSON input")
	data := map[string]interface{}{
		"deposit_address": hexutil.Encode(utils.RandomSlice(20)),
	}

	d, err := json.Marshal(data)
	assert.NoError(t, err)
	w, r = getHTTPReqAndResp(ctx, bytes.NewReader(d))
	srv.On("MintNFT", ctx, mock.Anything).Return(nil, nil, errors.New("failed to mint nft")).Once()
	h.MintInvoiceUnpaidNFT(w, r)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "failed to mint nft")
	srv.AssertExpectations(t)

	// success
	w, r = getHTTPReqAndResp(ctx, bytes.NewReader(d))
	tokenID := hexutil.Encode(utils.RandomSlice(32))
	srv.On("MintNFT", ctx, mock.Anything).Return(
		&nft.TokenResponse{
			TokenID: tokenID,
			JobID:   jobs.NewJobID().String(),
		}, nil, nil).Once()
	h.MintInvoiceUnpaidNFT(w, r)
	assert.Equal(t, http.StatusAccepted, w.Code)
	assert.Contains(t, w.Body.String(), "job_id")
	srv.AssertExpectations(t)
}
