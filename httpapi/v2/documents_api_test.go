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

	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/documents/generic"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/httpapi/coreapi"
	"github.com/centrifuge/go-centrifuge/pending"
	testingdocuments "github.com/centrifuge/go-centrifuge/testingutils/documents"
	testingidentity "github.com/centrifuge/go-centrifuge/testingutils/identity"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/go-chi/chi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func documentData() map[string]interface{} {
	return map[string]interface{}{}
}

func invalidDocIDPayload(t *testing.T) io.Reader {
	p := map[string]interface{}{
		"scheme":      "generic",
		"data":        documentData(),
		"document_id": "invalid",
	}

	d, err := json.Marshal(p)
	assert.NoError(t, err)
	return bytes.NewReader(d)
}

func validPayload(t *testing.T) io.Reader {
	p := map[string]interface{}{
		"scheme":      "generic",
		"data":        documentData(),
		"document_id": hexutil.Encode(utils.RandomSlice(32)),
		"attributes": map[string]map[string]string{
			"string_test": {
				"type":  "string",
				"value": "hello, world",
			},
		},
	}

	d, err := json.Marshal(p)
	assert.NoError(t, err)
	return bytes.NewReader(d)
}

func validClonePayload(t *testing.T) io.Reader {
	p := map[string]interface{}{
		"scheme": "generic",
	}

	d, err := json.Marshal(p)
	assert.NoError(t, err)
	return bytes.NewReader(d)
}

func invalidClonePayload(t *testing.T) io.Reader {
	p := map[string]interface{}{
		"scheme": "something_random",
	}

	d, err := json.Marshal(p)
	assert.NoError(t, err)
	return bytes.NewReader(d)
}

func invalidAttrPayload(t *testing.T) io.Reader {
	p := map[string]interface{}{
		"scheme": "generic",
		"data":   documentData(),
		"attributes": map[string]map[string]string{
			"string_test": {
				"type":  "invalid",
				"value": "hello, world",
			},
		},
	}

	d, err := json.Marshal(p)
	assert.NoError(t, err)
	return bytes.NewReader(d)
}

func TestHandler_CreateDocument(t *testing.T) {
	getHTTPReqAndResp := func(ctx context.Context, b io.Reader) (*httptest.ResponseRecorder, *http.Request) {
		return httptest.NewRecorder(), httptest.NewRequest("POST", "/documents", b).WithContext(ctx)
	}

	// failed unmarshal empty body
	ctx := context.Background()
	w, r := getHTTPReqAndResp(ctx, nil)
	pendingSrv := new(pending.MockService)
	h := handler{srv: Service{pendingDocSrv: pendingSrv}}
	h.CreateDocument(w, r)
	assert.Equal(t, w.Code, http.StatusBadRequest)
	assert.Contains(t, w.Body.String(), "unexpected end of JSON input")

	// failed unmarshal invalid doc_id
	w, r = getHTTPReqAndResp(ctx, invalidDocIDPayload(t))
	h.CreateDocument(w, r)
	assert.Equal(t, w.Code, http.StatusBadRequest)
	assert.Contains(t, w.Body.String(), "hex string without 0x prefix")

	// failed payloadConversion
	w, r = getHTTPReqAndResp(ctx, invalidAttrPayload(t))
	h.CreateDocument(w, r)
	assert.Equal(t, w.Code, http.StatusBadRequest)
	assert.Contains(t, w.Body.String(), "not a valid attribute type")

	// failed to create document
	pendingSrv.On("Create", ctx, mock.Anything).Return(nil, errors.New("Failed to create document")).Once()
	w, r = getHTTPReqAndResp(ctx, validPayload(t))
	h.CreateDocument(w, r)
	assert.Equal(t, w.Code, http.StatusBadRequest)
	assert.Contains(t, w.Body.String(), "Failed to create document")

	// failed document conversion
	doc := new(testingdocuments.MockModel)
	doc.On("GetData").Return(generic.Data{}).Twice()
	doc.On("Scheme").Return("generic").Twice()
	doc.On("GetAttributes").Return(nil).Twice()
	doc.On("GetCollaborators", mock.Anything).Return(documents.CollaboratorsAccess{}, errors.New("failed to get collaborators")).Once()
	pendingSrv.On("Create", ctx, mock.Anything).Return(doc, nil)
	w, r = getHTTPReqAndResp(ctx, validPayload(t))
	h.CreateDocument(w, r)
	assert.Equal(t, w.Code, http.StatusInternalServerError)
	assert.Contains(t, w.Body.String(), "failed to get collaborators")

	// success
	doc.On("GetCollaborators", mock.Anything).Return(documents.CollaboratorsAccess{}, nil).Once()
	doc.On("ID").Return(utils.RandomSlice(32)).Once()
	doc.On("CurrentVersion").Return(utils.RandomSlice(32)).Once()
	doc.On("Author").Return(nil, errors.New("somerror")).Once()
	doc.On("Timestamp").Return(nil, errors.New("somerror")).Once()
	doc.On("NFTs").Return(nil).Once()
	doc.On("GetStatus").Return(documents.Pending).Once()
	doc.On("CalculateTransitionRulesFingerprint").Return(utils.RandomSlice(32), nil)
	w, r = getHTTPReqAndResp(ctx, validPayload(t))
	h.CreateDocument(w, r)
	assert.Equal(t, w.Code, http.StatusCreated)
	assert.Contains(t, w.Body.String(), "\"status\":\"pending\"")
	pendingSrv.AssertExpectations(t)
	doc.AssertExpectations(t)
}

func TestHandler_CloneDocument(t *testing.T) {
	getHTTPReqAndResp := func(ctx context.Context, b io.Reader) (*httptest.ResponseRecorder, *http.Request) {
		return httptest.NewRecorder(), httptest.NewRequest("POST", "/documents/{document_id}/clone", b).WithContext(ctx)
	}

	// empty document_id
	rctx := chi.NewRouteContext()
	rctx.URLParams.Keys = make([]string, 1, 1)
	rctx.URLParams.Values = make([]string, 1, 1)
	rctx.URLParams.Keys[0] = "document_id"
	rctx.URLParams.Values[0] = ""
	ctx := context.WithValue(context.Background(), chi.RouteCtxKey, rctx)
	w, r := getHTTPReqAndResp(ctx, nil)
	h := handler{}
	h.CloneDocument(w, r)
	assert.Equal(t, w.Code, http.StatusBadRequest)
	assert.Contains(t, w.Body.String(), coreapi.ErrInvalidDocumentID.Error())

	// failed unmarshal empty body
	rctx.URLParams.Values[0] = hexutil.Encode(utils.RandomSlice(32))
	pendingSrv := new(pending.MockService)
	h = handler{srv: Service{pendingDocSrv: pendingSrv}}
	h.CloneDocument(w, r)
	assert.Equal(t, w.Code, http.StatusBadRequest)
	assert.Contains(t, w.Body.String(), "unexpected end of JSON input")

	// success
	w, r = getHTTPReqAndResp(ctx, validPayload(t))
	doc := new(testingdocuments.MockModel)
	doc.On("GetData").Return(generic.Data{})
	doc.On("Scheme").Return("generic")
	doc.On("GetAttributes").Return(nil)
	doc.On("GetCollaborators", mock.Anything).Return(documents.CollaboratorsAccess{}, nil).Once()
	doc.On("ID").Return(utils.RandomSlice(32)).Once()
	doc.On("CurrentVersion").Return(utils.RandomSlice(32)).Once()
	doc.On("Author").Return(nil, errors.New("somerror")).Once()
	doc.On("Timestamp").Return(nil, errors.New("somerror")).Once()
	doc.On("NFTs").Return(nil).Once()
	doc.On("GetStatus").Return(documents.Pending).Once()
	doc.On("CalculateTransitionRulesFingerprint").Return(utils.RandomSlice(32), nil)

	pendingSrv.On("Clone", ctx, mock.Anything, mock.Anything).Return(doc, nil)
	w, r = getHTTPReqAndResp(ctx, validClonePayload(t))

	h.CloneDocument(w, r)
	assert.Equal(t, w.Code, http.StatusCreated)
	assert.Contains(t, w.Body.String(), "\"status\":\"pending\"")
	pendingSrv.AssertExpectations(t)
	doc.AssertExpectations(t)
}

func TestHandler_UpdateDocument(t *testing.T) {
	getHTTPReqAndResp := func(ctx context.Context, b io.Reader) (*httptest.ResponseRecorder, *http.Request) {
		return httptest.NewRecorder(), httptest.NewRequest("PATCH", "/documents/{document_id}", b).WithContext(ctx)
	}

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
	assert.Contains(t, w.Body.String(), coreapi.ErrInvalidDocumentID.Error())

	// invalid id
	rctx.URLParams.Values[0] = "some invalid id"
	w, r = getHTTPReqAndResp(ctx, nil)
	h.UpdateDocument(w, r)
	assert.Equal(t, w.Code, http.StatusBadRequest)
	assert.Contains(t, w.Body.String(), coreapi.ErrInvalidDocumentID.Error())

	// failed unmarshal empty body
	rctx.URLParams.Values[0] = hexutil.Encode(utils.RandomSlice(32))
	w, r = getHTTPReqAndResp(ctx, nil)
	pendingSrv := new(pending.MockService)
	h = handler{srv: Service{pendingDocSrv: pendingSrv}}
	h.UpdateDocument(w, r)
	assert.Equal(t, w.Code, http.StatusBadRequest)
	assert.Contains(t, w.Body.String(), "unexpected end of JSON input")

	// failed payloadConversion
	w, r = getHTTPReqAndResp(ctx, invalidAttrPayload(t))
	h.UpdateDocument(w, r)
	assert.Equal(t, w.Code, http.StatusBadRequest)
	assert.Contains(t, w.Body.String(), "not a valid attribute type")

	// failed to update document
	pendingSrv.On("Update", ctx, mock.Anything).Return(nil, errors.New("Failed to update document")).Once()
	w, r = getHTTPReqAndResp(ctx, validPayload(t))
	h.UpdateDocument(w, r)
	assert.Equal(t, w.Code, http.StatusBadRequest)
	assert.Contains(t, w.Body.String(), "Failed to update document")

	// failed document conversion
	doc := new(testingdocuments.MockModel)
	doc.On("GetData").Return(generic.Data{}).Twice()
	doc.On("Scheme").Return("generic").Twice()
	doc.On("GetAttributes").Return(nil).Twice()
	doc.On("GetCollaborators", mock.Anything).Return(documents.CollaboratorsAccess{}, errors.New("failed to get collaborators")).Once()
	pendingSrv.On("Update", ctx, mock.Anything).Return(doc, nil)
	w, r = getHTTPReqAndResp(ctx, validPayload(t))
	h.UpdateDocument(w, r)
	assert.Equal(t, w.Code, http.StatusInternalServerError)
	assert.Contains(t, w.Body.String(), "failed to get collaborators")

	// success
	doc.On("GetCollaborators", mock.Anything).Return(documents.CollaboratorsAccess{}, nil).Once()
	doc.On("ID").Return(utils.RandomSlice(32)).Once()
	doc.On("CurrentVersion").Return(utils.RandomSlice(32)).Once()
	doc.On("Author").Return(nil, errors.New("somerror")).Once()
	doc.On("Timestamp").Return(nil, errors.New("somerror")).Once()
	doc.On("NFTs").Return(nil).Once()
	doc.On("GetStatus").Return(documents.Pending).Once()
	doc.On("CalculateTransitionRulesFingerprint").Return(utils.RandomSlice(32), nil)
	w, r = getHTTPReqAndResp(ctx, validPayload(t))
	h.UpdateDocument(w, r)
	assert.Equal(t, w.Code, http.StatusOK)
	assert.Contains(t, w.Body.String(), "\"status\":\"pending\"")
	pendingSrv.AssertExpectations(t)
	doc.AssertExpectations(t)
}

func TestHandler_Commit(t *testing.T) {
	getHTTPReqAndResp := func(ctx context.Context, b io.Reader) (*httptest.ResponseRecorder, *http.Request) {
		return httptest.NewRecorder(), httptest.NewRequest("POST", "/documents/{document_id}/commit", b).WithContext(ctx)
	}

	// empty document_id
	rctx := chi.NewRouteContext()
	rctx.URLParams.Keys = make([]string, 1, 1)
	rctx.URLParams.Values = make([]string, 1, 1)
	rctx.URLParams.Keys[0] = "document_id"
	rctx.URLParams.Values[0] = ""
	ctx := context.WithValue(context.Background(), chi.RouteCtxKey, rctx)
	w, r := getHTTPReqAndResp(ctx, nil)
	h := handler{}
	h.Commit(w, r)
	assert.Equal(t, w.Code, http.StatusBadRequest)
	assert.Contains(t, w.Body.String(), coreapi.ErrInvalidDocumentID.Error())

	// invalid hex
	rctx.URLParams.Values[0] = "invalid hex"
	w, r = getHTTPReqAndResp(ctx, nil)
	h.Commit(w, r)
	assert.Equal(t, w.Code, http.StatusBadRequest)
	assert.Contains(t, w.Body.String(), coreapi.ErrInvalidDocumentID.Error())

	// commit error
	rctx.URLParams.Values[0] = hexutil.Encode(utils.RandomSlice(32))
	srv := new(pending.MockService)
	h = handler{srv: Service{pendingDocSrv: srv}}
	srv.On("Commit", ctx, mock.Anything).Return(nil, nil, errors.New("Failed to commit document")).Once()
	w, r = getHTTPReqAndResp(ctx, nil)
	h.Commit(w, r)
	assert.Equal(t, w.Code, http.StatusBadRequest)
	assert.Contains(t, w.Body.String(), "Failed to commit document")

	// failed to convert collaborators in document
	doc := new(testingdocuments.MockModel)
	doc.On("GetData").Return(generic.Data{}).Twice()
	doc.On("Scheme").Return("generic").Twice()
	doc.On("GetAttributes").Return(nil).Twice()
	doc.On("GetCollaborators", mock.Anything).Return(documents.CollaboratorsAccess{}, errors.New("failed to get collaborators")).Once()
	srv.On("Commit", ctx, mock.Anything).Return(doc, nil, nil)
	w, r = getHTTPReqAndResp(ctx, nil)
	h.Commit(w, r)
	assert.Equal(t, w.Code, http.StatusInternalServerError)
	assert.Contains(t, w.Body.String(), "failed to get collaborators")

	// success
	doc.On("GetCollaborators", mock.Anything).Return(documents.CollaboratorsAccess{}, nil).Once()
	doc.On("ID").Return(utils.RandomSlice(32)).Once()
	doc.On("CurrentVersion").Return(utils.RandomSlice(32)).Once()
	doc.On("Author").Return(nil, errors.New("somerror")).Once()
	doc.On("Timestamp").Return(nil, errors.New("somerror")).Once()
	doc.On("NFTs").Return(nil).Once()
	doc.On("GetStatus").Return(documents.Committing).Once()
	doc.On("CalculateTransitionRulesFingerprint").Return(utils.RandomSlice(32), nil)
	w, r = getHTTPReqAndResp(ctx, validPayload(t))
	h.Commit(w, r)
	assert.Equal(t, http.StatusAccepted, w.Code)
	assert.Contains(t, w.Body.String(), "\"status\":\"committing\"")
	srv.AssertExpectations(t)
	doc.AssertExpectations(t)
}

func TestHandler_GetDocument(t *testing.T) {
	getHTTPReqAndResp := func(ctx context.Context, b io.Reader) (*httptest.ResponseRecorder, *http.Request) {
		return httptest.NewRecorder(), httptest.NewRequest("GET", "/documents/{document_id}/pending", b).WithContext(ctx)
	}

	// invalid id
	rctx := chi.NewRouteContext()
	rctx.URLParams.Keys = make([]string, 1, 1)
	rctx.URLParams.Values = make([]string, 1, 1)
	rctx.URLParams.Keys[0] = "document_id"
	rctx.URLParams.Values[0] = "some invalid id"
	ctx := context.WithValue(context.Background(), chi.RouteCtxKey, rctx)
	w, r := getHTTPReqAndResp(ctx, nil)
	h := handler{}
	h.getDocumentWithStatus(w, r, documents.Pending)
	assert.Equal(t, w.Code, http.StatusBadRequest)
	assert.Contains(t, w.Body.String(), coreapi.ErrInvalidDocumentID.Error())

	// missing document
	docID := utils.RandomSlice(32)
	rctx.URLParams.Values[0] = hexutil.Encode(docID)
	pendingSrv := new(pending.MockService)
	pendingSrv.On("Get", ctx, docID, documents.Pending).Return(nil, coreapi.ErrDocumentNotFound).Once()
	h.srv.pendingDocSrv = pendingSrv
	w, r = getHTTPReqAndResp(ctx, nil)
	h.getDocumentWithStatus(w, r, documents.Pending)
	assert.Equal(t, w.Code, http.StatusNotFound)
	assert.Contains(t, w.Body.String(), coreapi.ErrDocumentNotFound.Error())

	// failed conversion
	doc := new(testingdocuments.MockModel)
	doc.On("GetData").Return(generic.Data{}).Times(3)
	doc.On("Scheme").Return("generic").Times(3)
	doc.On("GetAttributes").Return(nil).Times(3)
	doc.On("GetCollaborators", mock.Anything).Return(documents.CollaboratorsAccess{}, errors.New("failed to get collaborators")).Once()
	pendingSrv.On("Get", ctx, docID, mock.Anything).Return(doc, nil)
	w, r = getHTTPReqAndResp(ctx, nil)
	h.getDocumentWithStatus(w, r, documents.Pending)
	assert.Equal(t, w.Code, http.StatusInternalServerError)
	assert.Contains(t, w.Body.String(), "failed to get collaborators")

	// success pending
	doc.On("GetCollaborators", mock.Anything).Return(documents.CollaboratorsAccess{}, nil).Twice()
	doc.On("ID").Return(utils.RandomSlice(32)).Twice()
	doc.On("CurrentVersion").Return(utils.RandomSlice(32)).Twice()
	doc.On("Author").Return(nil, errors.New("somerror")).Twice()
	doc.On("Timestamp").Return(nil, errors.New("somerror")).Twice()
	doc.On("NFTs").Return(nil).Twice()
	doc.On("GetStatus").Return(documents.Pending).Twice()
	doc.On("CalculateTransitionRulesFingerprint").Return(utils.RandomSlice(32), nil)
	w, r = getHTTPReqAndResp(ctx, nil)
	h.GetPendingDocument(w, r)
	assert.Equal(t, http.StatusOK, w.Code)
	w, r = getHTTPReqAndResp(ctx, nil)
	h.GetCommittedDocument(w, r)
	assert.Equal(t, http.StatusOK, w.Code)
	pendingSrv.AssertExpectations(t)
	doc.AssertExpectations(t)
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
		assert.Contains(t, w.Body.String(), coreapi.ErrInvalidDocumentID.Error())
	}

	// missing document
	docID := utils.RandomSlice(32)
	versionID := utils.RandomSlice(32)
	rctx.URLParams.Values[0] = hexutil.Encode(docID)
	rctx.URLParams.Values[1] = hexutil.Encode(versionID)
	pendingSrv := new(pending.MockService)
	pendingSrv.On("GetVersion", ctx, docID, versionID).Return(nil, coreapi.ErrDocumentNotFound).Once()
	h.srv.pendingDocSrv = pendingSrv
	w, r := getHTTPReqAndResp(ctx)
	h.GetDocumentVersion(w, r)
	assert.Equal(t, w.Code, http.StatusNotFound)
	assert.Contains(t, w.Body.String(), coreapi.ErrDocumentNotFound.Error())

	// failed conversion
	doc := new(testingdocuments.MockModel)
	doc.On("GetData").Return(generic.Data{}).Times(2)
	doc.On("Scheme").Return("generic").Times(2)
	doc.On("GetAttributes").Return(nil).Times(2)
	doc.On("GetCollaborators", mock.Anything).Return(documents.CollaboratorsAccess{}, errors.New("failed to get collaborators")).Once()
	pendingSrv.On("GetVersion", ctx, docID, versionID).Return(doc, nil)
	w, r = getHTTPReqAndResp(ctx)
	h.GetDocumentVersion(w, r)
	assert.Equal(t, w.Code, http.StatusInternalServerError)
	assert.Contains(t, w.Body.String(), "failed to get collaborators")

	// success pending
	doc.On("GetCollaborators", mock.Anything).Return(documents.CollaboratorsAccess{}, nil).Once()
	doc.On("ID").Return(utils.RandomSlice(32)).Once()
	doc.On("CurrentVersion").Return(utils.RandomSlice(32)).Once()
	doc.On("Author").Return(nil, errors.New("somerror")).Once()
	doc.On("Timestamp").Return(nil, errors.New("somerror")).Once()
	doc.On("NFTs").Return(nil).Once()
	doc.On("GetStatus").Return(documents.Pending).Once()
	doc.On("CalculateTransitionRulesFingerprint").Return(utils.RandomSlice(32), nil)
	w, r = getHTTPReqAndResp(ctx)
	h.GetDocumentVersion(w, r)
	assert.Equal(t, http.StatusOK, w.Code)
	pendingSrv.AssertExpectations(t)
	doc.AssertExpectations(t)
}

func TestHandler_RemoveCollaborators(t *testing.T) {
	getHTTPReqAndResp := func(ctx context.Context, b io.Reader) (*httptest.ResponseRecorder, *http.Request) {
		return httptest.NewRecorder(), httptest.NewRequest("DELETE", "/documents/{document_id}/collaborators", b).WithContext(ctx)
	}

	// empty document_id
	rctx := chi.NewRouteContext()
	rctx.URLParams.Keys = make([]string, 1, 1)
	rctx.URLParams.Values = make([]string, 1, 1)
	rctx.URLParams.Keys[0] = "document_id"
	rctx.URLParams.Values[0] = ""
	ctx := context.WithValue(context.Background(), chi.RouteCtxKey, rctx)
	w, r := getHTTPReqAndResp(ctx, nil)
	h := handler{}
	h.RemoveCollaborators(w, r)
	assert.Equal(t, w.Code, http.StatusBadRequest)
	assert.Contains(t, w.Body.String(), coreapi.ErrInvalidDocumentID.Error())

	// invalid id
	rctx.URLParams.Values[0] = "some invalid id"
	w, r = getHTTPReqAndResp(ctx, nil)
	h.RemoveCollaborators(w, r)
	assert.Equal(t, w.Code, http.StatusBadRequest)
	assert.Contains(t, w.Body.String(), coreapi.ErrInvalidDocumentID.Error())

	// failed unmarshal empty body
	docID := utils.RandomSlice(32)
	rctx.URLParams.Values[0] = hexutil.Encode(docID)
	w, r = getHTTPReqAndResp(ctx, nil)
	pendingSrv := new(pending.MockService)
	h = handler{srv: Service{pendingDocSrv: pendingSrv}}
	h.RemoveCollaborators(w, r)
	assert.Equal(t, w.Code, http.StatusBadRequest)
	assert.Contains(t, w.Body.String(), "unexpected end of JSON input")

	// failed to remove collaborators
	req := map[string]interface{}{
		"collaborators": []string{testingidentity.GenerateRandomDID().String()},
	}
	d, err := json.Marshal(req)
	assert.NoError(t, err)
	pendingSrv.On("RemoveCollaborators", ctx, docID, mock.Anything).Return(nil, errors.New("failed to delete collaborators")).Once()
	w, r = getHTTPReqAndResp(ctx, bytes.NewReader(d))
	h.RemoveCollaborators(w, r)
	assert.Equal(t, w.Code, http.StatusBadRequest)
	assert.Contains(t, w.Body.String(), "failed to delete collaborators")

	// failed conversion
	doc := new(testingdocuments.MockModel)
	doc.On("GetData").Return(generic.Data{}).Twice()
	doc.On("Scheme").Return("generic").Twice()
	doc.On("GetAttributes").Return(nil).Twice()
	doc.On("GetCollaborators", mock.Anything).Return(documents.CollaboratorsAccess{}, errors.New("failed to get collaborators")).Once()
	pendingSrv.On("RemoveCollaborators", ctx, docID, mock.Anything).Return(doc, nil)
	w, r = getHTTPReqAndResp(ctx, bytes.NewReader(d))
	h.RemoveCollaborators(w, r)
	assert.Equal(t, w.Code, http.StatusInternalServerError)
	assert.Contains(t, w.Body.String(), "failed to get collaborators")

	// success
	doc.On("GetCollaborators", mock.Anything).Return(documents.CollaboratorsAccess{}, nil).Once()
	doc.On("ID").Return(utils.RandomSlice(32)).Once()
	doc.On("CurrentVersion").Return(utils.RandomSlice(32)).Once()
	doc.On("Author").Return(nil, errors.New("somerror")).Once()
	doc.On("Timestamp").Return(nil, errors.New("somerror")).Once()
	doc.On("NFTs").Return(nil).Once()
	doc.On("GetStatus").Return(documents.Pending).Once()
	doc.On("CalculateTransitionRulesFingerprint").Return(utils.RandomSlice(32), nil)
	w, r = getHTTPReqAndResp(ctx, bytes.NewReader(d))
	h.RemoveCollaborators(w, r)
	assert.Equal(t, w.Code, http.StatusOK)
	doc.AssertExpectations(t)
	pendingSrv.AssertExpectations(t)
}
