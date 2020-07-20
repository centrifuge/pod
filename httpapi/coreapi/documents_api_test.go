// +build unit

package coreapi

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
	"github.com/centrifuge/precise-proofs/proofs"
	"github.com/centrifuge/precise-proofs/proofs/proto"
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
	srv := Service{docSrv: docSrv}
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
	srv = Service{docSrv: docSrv}
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
	m.On("CalculateTransitionRulesFingerprint").Return(utils.RandomSlice(32), nil)
	docSrv = new(testingdocuments.MockService)
	srv = Service{docSrv: docSrv}
	h = handler{srv: srv}
	docSrv.On("CreateModel", mock.Anything, mock.Anything).Return(m, jobs.NewJobID(), nil)
	r = httptest.NewRequest("POST", "/documents", bytes.NewReader(d))
	w = httptest.NewRecorder()
	h.CreateDocument(w, r)
	assert.Equal(t, w.Code, http.StatusAccepted)
	m.AssertExpectations(t)
	docSrv.AssertExpectations(t)
}

func TestHandler_UpdateDocument(t *testing.T) {
	getHTTPReqAndResp := func(ctx context.Context, b io.Reader) (*httptest.ResponseRecorder, *http.Request) {
		return httptest.NewRecorder(), httptest.NewRequest("PUT", "/documents/{document_id}", b).WithContext(ctx)
	}

	m := new(testingdocuments.MockModel)

	// empty document_id
	rctx := chi.NewRouteContext()
	rctx.URLParams.Keys = make([]string, 1, 1)
	rctx.URLParams.Values = make([]string, 1, 1)
	rctx.URLParams.Keys[0] = "document_id"
	rctx.URLParams.Values[0] = ""
	ctx := context.WithValue(context.Background(), chi.RouteCtxKey, rctx)
	w, r := getHTTPReqAndResp(ctx, nil)
	h := handler{}
	h.UpdateDocument(w, r)
	assert.Equal(t, w.Code, http.StatusBadRequest)
	assert.Contains(t, w.Body.String(), ErrInvalidDocumentID.Error())

	// invalid id
	rctx.URLParams.Values[0] = "some invalid id"
	w, r = getHTTPReqAndResp(ctx, nil)
	h.UpdateDocument(w, r)
	assert.Equal(t, w.Code, http.StatusBadRequest)
	assert.Contains(t, w.Body.String(), ErrInvalidDocumentID.Error())
	id := hexutil.Encode(utils.RandomSlice(32))
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
	rctx.URLParams.Values[0] = id
	w, r = getHTTPReqAndResp(ctx, bytes.NewReader(d))
	h.UpdateDocument(w, r)
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
	srv := Service{docSrv: docSrv}
	h = handler{srv: srv}
	docSrv.On("UpdateModel", mock.Anything, mock.Anything).Return(nil, jobs.NilJobID(), errors.New("failed to update model"))
	w, r = getHTTPReqAndResp(ctx, bytes.NewReader(d))
	h.UpdateDocument(w, r)
	assert.Equal(t, w.Code, http.StatusBadRequest)
	assert.Contains(t, w.Body.String(), "failed to update model")
	docSrv.AssertExpectations(t)

	m.On("GetData").Return(data)
	m.On("Scheme").Return("invoice")
	m.On("GetAttributes").Return(nil)
	m.On("GetCollaborators", mock.Anything).Return(documents.CollaboratorsAccess{}, errors.New("failed to get collaborators"))
	docSrv = new(testingdocuments.MockService)
	srv = Service{docSrv: docSrv}
	h = handler{srv: srv}
	docSrv.On("UpdateModel", mock.Anything, mock.Anything).Return(m, jobs.NewJobID(), nil)
	w, r = getHTTPReqAndResp(ctx, bytes.NewReader(d))
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
	m.On("CalculateTransitionRulesFingerprint").Return(utils.RandomSlice(32), nil)
	docSrv = new(testingdocuments.MockService)
	srv = Service{docSrv: docSrv}
	h = handler{srv: srv}
	docSrv.On("UpdateModel", mock.Anything, mock.Anything).Return(m, jobs.NewJobID(), nil)
	w, r = getHTTPReqAndResp(ctx, bytes.NewReader(d))
	h.UpdateDocument(w, r)
	assert.Equal(t, w.Code, http.StatusAccepted)
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
	h = handler{srv: Service{docSrv: docSrv}}
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
	h = handler{srv: Service{docSrv: docSrv}}
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
	m.On("CalculateTransitionRulesFingerprint").Return(utils.RandomSlice(32), nil)
	docSrv = new(testingdocuments.MockService)
	docSrv.On("GetCurrentVersion", id).Return(m, nil)
	h = handler{srv: Service{docSrv: docSrv}}
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
	h = handler{srv: Service{docSrv: docSrv}}
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
	h = handler{srv: Service{docSrv: docSrv}}
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
	m.On("CalculateTransitionRulesFingerprint").Return(utils.RandomSlice(32), nil)
	docSrv = new(testingdocuments.MockService)
	docSrv.On("GetVersion", id, vid).Return(m, nil)
	h = handler{srv: Service{docSrv: docSrv}}
	w, r = getHTTPReqAndResp(ctx)
	h.GetDocumentVersion(w, r)
	assert.Equal(t, w.Code, http.StatusOK)
	docSrv.AssertExpectations(t)
	m.AssertExpectations(t)
}

func TestHandler_GenerateProofs(t *testing.T) {
	getHTTPReqAndResp := func(ctx context.Context, body io.Reader) (*httptest.ResponseRecorder, *http.Request) {
		return httptest.NewRecorder(), httptest.NewRequest("GET", "/documents/{document_id}/proofs", body).WithContext(ctx)
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
		w, r := getHTTPReqAndResp(ctx, nil)
		h.GenerateProofs(w, r)
		assert.Equal(t, w.Code, http.StatusBadRequest)
		assert.Contains(t, w.Body.String(), ErrInvalidDocumentID.Error())
	}

	// failed json input
	id := utils.RandomSlice(32)
	rctx.URLParams.Values[0] = hexutil.Encode(id)
	h = handler{}
	w, r := getHTTPReqAndResp(ctx, nil)
	h.GenerateProofs(w, r)
	assert.Equal(t, w.Code, http.StatusBadRequest)
	assert.Contains(t, w.Body.String(), "unexpected end of JSON input")

	// failed to generate proofs
	request := ProofsRequest{}
	d, err := json.Marshal(request)
	assert.NoError(t, err)
	buf := bytes.NewReader(d)
	docSrv := new(testingdocuments.MockService)
	docSrv.On("CreateProofs", mock.Anything, id, request.Fields).Return(nil, errors.New("failed to generate proofs"))
	h = handler{srv: Service{docSrv: docSrv}}
	w, r = getHTTPReqAndResp(ctx, buf)
	h.GenerateProofs(w, r)
	assert.Equal(t, w.Code, http.StatusInternalServerError)
	assert.Contains(t, w.Body.String(), "failed to generate proofs")
	docSrv.AssertExpectations(t)

	// success
	buf = bytes.NewReader(d)
	v1, err := hexutil.Decode("0x76616c756531")
	assert.NoError(t, err)
	proof := &documents.DocumentProof{
		DocumentID: id,
		VersionID:  id,
		State:      "state",
		FieldProofs: []*proofspb.Proof{
			{
				Property: proofs.CompactName([]byte{0, 0, 1}...),
				Value:    v1,
				Salt:     []byte{1, 2, 3},
				Hash:     []byte{1, 2, 4},
				SortedHashes: [][]byte{
					{1, 2, 5},
					{1, 2, 6},
					{1, 2, 7},
				},
			},
		},
	}
	docSrv = new(testingdocuments.MockService)
	docSrv.On("CreateProofs", mock.Anything, id, request.Fields).Return(proof, nil)
	h = handler{srv: Service{docSrv: docSrv}}
	w, r = getHTTPReqAndResp(ctx, buf)
	h.GenerateProofs(w, r)
	assert.Equal(t, w.Code, http.StatusOK)
	assert.Contains(t, w.Body.String(), hexutil.Encode(id))
	docSrv.AssertExpectations(t)
}

func TestHandler_GenerateProofsForVersion(t *testing.T) {
	getHTTPReqAndResp := func(ctx context.Context, body io.Reader) (*httptest.ResponseRecorder, *http.Request) {
		return httptest.NewRecorder(), httptest.NewRequest("GET", "/documents/{document_id}/versions/{version_id}/proofs", body).WithContext(ctx)
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
		w, r := getHTTPReqAndResp(ctx, nil)
		h.GenerateProofsForVersion(w, r)
		assert.Equal(t, w.Code, http.StatusBadRequest)
		assert.Contains(t, w.Body.String(), ErrInvalidDocumentID.Error())
	}

	// failed json input
	id := utils.RandomSlice(32)
	vid := utils.RandomSlice(32)
	rctx.URLParams.Values[0] = hexutil.Encode(id)
	rctx.URLParams.Values[1] = hexutil.Encode(vid)
	h = handler{}
	w, r := getHTTPReqAndResp(ctx, nil)
	h.GenerateProofsForVersion(w, r)
	assert.Equal(t, w.Code, http.StatusBadRequest)
	assert.Contains(t, w.Body.String(), "unexpected end of JSON input")

	// failed to generate proofs
	request := ProofsRequest{}
	d, err := json.Marshal(request)
	assert.NoError(t, err)
	buf := bytes.NewReader(d)
	docSrv := new(testingdocuments.MockService)
	docSrv.On("CreateProofsForVersion", mock.Anything, id, vid, request.Fields).Return(nil, errors.New("failed to generate proofs"))
	h = handler{srv: Service{docSrv: docSrv}}
	w, r = getHTTPReqAndResp(ctx, buf)
	h.GenerateProofsForVersion(w, r)
	assert.Equal(t, w.Code, http.StatusInternalServerError)
	assert.Contains(t, w.Body.String(), "failed to generate proofs")
	docSrv.AssertExpectations(t)

	// success
	buf = bytes.NewReader(d)
	docSrv = new(testingdocuments.MockService)
	v1, err := hexutil.Decode("0x76616c756531")
	assert.NoError(t, err)
	proof := &documents.DocumentProof{
		DocumentID: id,
		VersionID:  id,
		State:      "state",
		FieldProofs: []*proofspb.Proof{
			{
				Property: proofs.CompactName([]byte{0, 0, 1}...),
				Value:    v1,
				Salt:     []byte{1, 2, 3},
				Hash:     []byte{1, 2, 4},
				SortedHashes: [][]byte{
					{1, 2, 5},
					{1, 2, 6},
					{1, 2, 7},
				},
			},
		},
	}
	docSrv.On("CreateProofsForVersion", mock.Anything, id, vid, request.Fields).Return(proof, nil)
	h = handler{srv: Service{docSrv: docSrv}}
	w, r = getHTTPReqAndResp(ctx, buf)
	h.GenerateProofsForVersion(w, r)
	assert.Equal(t, w.Code, http.StatusOK)
	assert.Contains(t, w.Body.String(), hexutil.Encode(id))
	docSrv.AssertExpectations(t)
}
