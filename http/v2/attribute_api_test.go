//go:build unit

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
	"github.com/centrifuge/go-centrifuge/http/coreapi"
	"github.com/centrifuge/go-centrifuge/pending"
	testingdocuments "github.com/centrifuge/go-centrifuge/testingutils/documents"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/go-chi/chi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestHandler_AddSignedAttribute(t *testing.T) {
	getHTTPReqAndResp := func(ctx context.Context, b io.Reader) (*httptest.ResponseRecorder, *http.Request) {
		return httptest.NewRecorder(), httptest.NewRequest("POST", "/documents/{document_id}/signed_attribute", b).WithContext(ctx)
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
	h.AddSignedAttribute(w, r)
	assert.Equal(t, w.Code, http.StatusBadRequest)
	assert.Contains(t, w.Body.String(), coreapi.ErrInvalidDocumentID.Error())

	// invalid id
	rctx.URLParams.Values[0] = "some invalid id"
	w, r = getHTTPReqAndResp(ctx, nil)
	h.AddSignedAttribute(w, r)
	assert.Equal(t, w.Code, http.StatusBadRequest)
	assert.Contains(t, w.Body.String(), coreapi.ErrInvalidDocumentID.Error())

	// failed unmarshal empty body
	docID := utils.RandomSlice(32)
	rctx.URLParams.Values[0] = hexutil.Encode(docID)
	w, r = getHTTPReqAndResp(ctx, nil)
	pendingSrv := new(pending.MockService)
	h = handler{srv: Service{pendingDocSrv: pendingSrv}}
	h.AddSignedAttribute(w, r)
	assert.Equal(t, w.Code, http.StatusBadRequest)
	assert.Contains(t, w.Body.String(), "unexpected end of JSON input")

	// failed to add attribute
	label := "signed_attribute"
	payload := utils.RandomSlice(32)
	req := map[string]string{
		"label":   label,
		"type":    "bytes",
		"payload": hexutil.Encode(payload),
	}
	d, err := json.Marshal(req)
	assert.NoError(t, err)
	pendingSrv.On("AddSignedAttribute", ctx, docID, label, payload).Return(nil, errors.New("failed to add attribute")).Once()
	w, r = getHTTPReqAndResp(ctx, bytes.NewReader(d))
	h.AddSignedAttribute(w, r)
	assert.Equal(t, w.Code, http.StatusBadRequest)
	assert.Contains(t, w.Body.String(), "failed to add attribute")

	// failed conversion
	doc := new(testingdocuments.MockModel)
	doc.On("GetData").Return(generic.Data{}).Twice()
	doc.On("Scheme").Return("generic").Twice()
	doc.On("GetAttributes").Return(nil).Twice()
	doc.On("GetCollaborators", mock.Anything).Return(documents.CollaboratorsAccess{}, errors.New("failed to get collaborators")).Once()
	pendingSrv.On("AddSignedAttribute", ctx, docID, label, payload).Return(doc, nil)
	w, r = getHTTPReqAndResp(ctx, bytes.NewReader(d))
	h.AddSignedAttribute(w, r)
	assert.Equal(t, w.Code, http.StatusInternalServerError)
	assert.Contains(t, w.Body.String(), "failed to get collaborators")

	// success
	doc.On("GetCollaborators", mock.Anything).Return(documents.CollaboratorsAccess{}, nil).Once()
	doc.On("ID").Return(utils.RandomSlice(32)).Once()
	var prevID []byte = nil
	doc.On("PreviousVersion").Return(prevID).Once()
	doc.On("CurrentVersion").Return(utils.RandomSlice(32)).Once()
	doc.On("NextVersion").Return(utils.RandomSlice(32)).Once()
	doc.On("Author").Return(nil, errors.New("somerror")).Once()
	doc.On("Timestamp").Return(nil, errors.New("somerror")).Once()
	doc.On("NFTs").Return(nil).Once()
	doc.On("GetStatus").Return(documents.Pending).Once()
	w, r = getHTTPReqAndResp(ctx, bytes.NewReader(d))
	doc.On("CalculateTransitionRulesFingerprint").Return(utils.RandomSlice(32), nil)
	h.AddSignedAttribute(w, r)
	assert.Equal(t, w.Code, http.StatusOK)
	doc.AssertExpectations(t)
	pendingSrv.AssertExpectations(t)
}

func TestHandler_AddAttributes(t *testing.T) {
	getHTTPReqAndResp := func(ctx context.Context, b io.Reader) (*httptest.ResponseRecorder, *http.Request) {
		return httptest.NewRecorder(), httptest.NewRequest("POST", "/documents/{document_id}/attributes",
			b).WithContext(ctx)
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
	h.AddAttributes(w, r)
	assert.Equal(t, w.Code, http.StatusBadRequest)
	assert.Contains(t, w.Body.String(), coreapi.ErrInvalidDocumentID.Error())

	// invalid id
	rctx.URLParams.Values[0] = "some invalid id"
	w, r = getHTTPReqAndResp(ctx, nil)
	h.AddAttributes(w, r)
	assert.Equal(t, w.Code, http.StatusBadRequest)
	assert.Contains(t, w.Body.String(), coreapi.ErrInvalidDocumentID.Error())

	// failed unmarshal empty body
	docID := utils.RandomSlice(32)
	rctx.URLParams.Values[0] = hexutil.Encode(docID)
	w, r = getHTTPReqAndResp(ctx, nil)
	pendingSrv := new(pending.MockService)
	h = handler{srv: Service{pendingDocSrv: pendingSrv}}
	h.AddAttributes(w, r)
	assert.Equal(t, w.Code, http.StatusBadRequest)
	assert.Contains(t, w.Body.String(), "unexpected end of JSON input")

	// failed to add attribute
	label := "test"
	payload := utils.RandomSlice(32)
	req := coreapi.AttributeMapRequest{
		label: coreapi.AttributeRequest{
			Type:  "bytes",
			Value: hexutil.Encode(payload),
		},
	}
	d, err := json.Marshal(req)
	assert.NoError(t, err)
	pendingSrv.On("AddAttributes", ctx, docID, mock.Anything).Return(nil, errors.New("failed to add attribute")).Once()
	w, r = getHTTPReqAndResp(ctx, bytes.NewReader(d))
	h.AddAttributes(w, r)
	assert.Equal(t, w.Code, http.StatusBadRequest)
	assert.Contains(t, w.Body.String(), "failed to add attribute")

	// failed conversion
	doc := new(testingdocuments.MockModel)
	doc.On("GetData").Return(generic.Data{}).Twice()
	doc.On("Scheme").Return("generic").Twice()
	doc.On("GetAttributes").Return(nil).Twice()
	doc.On("GetCollaborators", mock.Anything).Return(documents.CollaboratorsAccess{}, errors.New("failed to get collaborators")).Once()
	pendingSrv.On("AddAttributes", ctx, docID, mock.Anything).Return(doc, nil)
	w, r = getHTTPReqAndResp(ctx, bytes.NewReader(d))
	h.AddAttributes(w, r)
	assert.Equal(t, w.Code, http.StatusInternalServerError)
	assert.Contains(t, w.Body.String(), "failed to get collaborators")

	// success
	doc.On("GetCollaborators", mock.Anything).Return(documents.CollaboratorsAccess{}, nil).Once()
	doc.On("ID").Return(utils.RandomSlice(32)).Once()
	var prevID []byte = nil
	doc.On("PreviousVersion").Return(prevID).Once()
	doc.On("CurrentVersion").Return(utils.RandomSlice(32)).Once()
	doc.On("NextVersion").Return(utils.RandomSlice(32)).Once()
	doc.On("Author").Return(nil, errors.New("somerror")).Once()
	doc.On("Timestamp").Return(nil, errors.New("somerror")).Once()
	doc.On("NFTs").Return(nil).Once()
	doc.On("GetStatus").Return(documents.Pending).Once()
	w, r = getHTTPReqAndResp(ctx, bytes.NewReader(d))
	doc.On("CalculateTransitionRulesFingerprint").Return(utils.RandomSlice(32), nil)
	h.AddAttributes(w, r)
	assert.Equal(t, w.Code, http.StatusOK)
	doc.AssertExpectations(t)
	pendingSrv.AssertExpectations(t)
}

func TestHandler_DeleteAttribute(t *testing.T) {
	getHTTPReqAndResp := func(ctx context.Context) (*httptest.ResponseRecorder, *http.Request) {
		return httptest.NewRecorder(), httptest.NewRequest("DELETE",
			"/documents/{document_id}/attributes/{attribute_key}", nil).WithContext(ctx)
	}

	// empty document_id
	rctx := chi.NewRouteContext()
	rctx.URLParams.Keys = make([]string, 2, 2)
	rctx.URLParams.Values = make([]string, 2, 2)
	rctx.URLParams.Keys[0] = "document_id"
	rctx.URLParams.Keys[1] = "attribute_key"
	rctx.URLParams.Values[0] = ""
	ctx := context.WithValue(context.Background(), chi.RouteCtxKey, rctx)
	w, r := getHTTPReqAndResp(ctx)
	h := handler{}
	h.DeleteAttribute(w, r)
	assert.Equal(t, w.Code, http.StatusBadRequest)
	assert.Contains(t, w.Body.String(), coreapi.ErrInvalidDocumentID.Error())

	// empty attributeKey
	docID := utils.RandomSlice(32)
	rctx.URLParams.Values[0] = hexutil.Encode(docID)
	w, r = getHTTPReqAndResp(ctx)
	h.DeleteAttribute(w, r)
	assert.Equal(t, w.Code, http.StatusBadRequest)
	assert.Contains(t, w.Body.String(), ErrInvalidAttributeKey.Error())

	// invalid attributeKey
	rctx.URLParams.Values[1] = "some attribute key"
	w, r = getHTTPReqAndResp(ctx)
	h.DeleteAttribute(w, r)
	assert.Equal(t, w.Code, http.StatusBadRequest)
	assert.Contains(t, w.Body.String(), ErrInvalidAttributeKey.Error())

	// failed to delete attribute
	rctx.URLParams.Values[1] = hexutil.Encode(utils.RandomSlice(32))
	pendingSrv := new(pending.MockService)
	pendingSrv.On("DeleteAttribute", mock.Anything, docID, mock.Anything).Return(nil,
		errors.New("failed to delete attribute")).Once()
	h = handler{srv: Service{pendingDocSrv: pendingSrv}}
	h.DeleteAttribute(w, r)
	assert.Equal(t, w.Code, http.StatusBadRequest)

	// failed conversion
	doc := new(testingdocuments.MockModel)
	doc.On("GetData").Return(generic.Data{}).Twice()
	doc.On("Scheme").Return("generic").Twice()
	doc.On("GetAttributes").Return(nil).Twice()
	doc.On("GetCollaborators", mock.Anything).Return(documents.CollaboratorsAccess{}, errors.New("failed to get collaborators")).Once()
	pendingSrv.On("DeleteAttribute", mock.Anything, docID, mock.Anything).Return(doc, nil).Twice()
	w, r = getHTTPReqAndResp(ctx)
	h.DeleteAttribute(w, r)
	assert.Equal(t, w.Code, http.StatusInternalServerError)
	assert.Contains(t, w.Body.String(), "failed to get collaborators")

	// success
	doc.On("GetCollaborators", mock.Anything).Return(documents.CollaboratorsAccess{}, nil).Once()
	doc.On("ID").Return(utils.RandomSlice(32)).Once()
	var prevID []byte = nil
	doc.On("PreviousVersion").Return(prevID).Once()
	doc.On("CurrentVersion").Return(utils.RandomSlice(32)).Once()
	doc.On("NextVersion").Return(utils.RandomSlice(32)).Once()
	doc.On("Author").Return(nil, errors.New("somerror")).Once()
	doc.On("Timestamp").Return(nil, errors.New("somerror")).Once()
	doc.On("NFTs").Return(nil).Once()
	doc.On("GetStatus").Return(documents.Pending).Once()
	w, r = getHTTPReqAndResp(ctx)
	doc.On("CalculateTransitionRulesFingerprint").Return(utils.RandomSlice(32), nil)
	h.DeleteAttribute(w, r)
	assert.Equal(t, w.Code, http.StatusOK)
	doc.AssertExpectations(t)
	pendingSrv.AssertExpectations(t)
}
