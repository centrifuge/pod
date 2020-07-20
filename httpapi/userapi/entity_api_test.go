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

	coredocumentpb "github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/documents/entity"
	"github.com/centrifuge/go-centrifuge/documents/entityrelationship"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/httpapi/coreapi"
	"github.com/centrifuge/go-centrifuge/jobs"
	testingdocuments "github.com/centrifuge/go-centrifuge/testingutils/documents"
	testingidentity "github.com/centrifuge/go-centrifuge/testingutils/identity"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/centrifuge/go-centrifuge/utils/byteutils"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/go-chi/chi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func entityData() map[string]interface{} {
	return map[string]interface{}{
		"identity":   "0xBAEb33a61f05e6F269f1c4b4CFF91A901B54DaF7",
		"legal_name": "John doe",
		"addresses": []map[string]interface{}{
			{
				"is_main": true,
				"country": "germany",
			},
		},
	}
}

func marshall(t *testing.T, data interface{}) []byte {
	d, err := json.Marshal(data)
	assert.NoError(t, err)
	return d
}

func TestHandler_CreateEntity(t *testing.T) {
	data := map[string]interface{}{
		"data": entityData(),
		"attributes": map[string]map[string]string{
			"string_test": {
				"type":  "invalid",
				"value": "hello, world",
			},
		},
	}

	getHTTPReqAndResp := func(ctx context.Context, b io.Reader) (*httptest.ResponseRecorder, *http.Request) {
		return httptest.NewRecorder(), httptest.NewRequest("POST", "/entities", b).WithContext(ctx)
	}

	// empty body
	rctx := chi.NewRouteContext()
	ctx := context.WithValue(context.Background(), chi.RouteCtxKey, rctx)
	h := handler{}
	w, r := getHTTPReqAndResp(ctx, nil)
	h.CreateEntity(w, r)
	assert.Equal(t, w.Code, http.StatusBadRequest)
	assert.Contains(t, w.Body.String(), "unexpected end of JSON input")

	// invalid body
	w, r = getHTTPReqAndResp(ctx, bytes.NewReader(marshall(t, data)))
	h.CreateEntity(w, r)
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
	m.On("GetData").Return(entity.Data{})
	m.On("Scheme").Return(entity.Scheme)
	m.On("GetAttributes").Return(nil)
	m.On("GetCollaborators", mock.Anything).Return(documents.CollaboratorsAccess{}, errors.New("failed to get collaborators"))
	docSrv := new(testingdocuments.MockService)
	srv := Service{coreAPISrv: newCoreAPIService(docSrv)}
	h = handler{srv: srv}
	docSrv.On("CreateModel", mock.Anything, mock.Anything).Return(m, jobs.NewJobID(), nil)
	w, r = getHTTPReqAndResp(ctx, bytes.NewReader(marshall(t, data)))
	h.CreateEntity(w, r)
	assert.Equal(t, w.Code, http.StatusInternalServerError)
	assert.Contains(t, w.Body.String(), "failed to get collaborators")
	m.AssertExpectations(t)
	docSrv.AssertExpectations(t)

	// success
	m = new(testingdocuments.MockModel)
	m.On("GetCollaborators", mock.Anything).Return(documents.CollaboratorsAccess{}, nil).Once()
	m.On("GetData").Return(entity.Data{})
	m.On("Scheme").Return(entity.Scheme)
	m.On("ID").Return(utils.RandomSlice(32)).Once()
	m.On("CurrentVersion").Return(utils.RandomSlice(32)).Once()
	m.On("Author").Return(nil, errors.New("somerror"))
	m.On("Timestamp").Return(nil, errors.New("somerror"))
	m.On("NFTs").Return(nil)
	m.On("GetAttributes").Return(nil)
	collab := testingidentity.GenerateRandomDID()
	m.On("IsDIDCollaborator", collab).Return(false, nil).Once()
	ctx = context.WithValue(ctx, config.AccountHeaderKey, collab.String())
	docSrv = new(testingdocuments.MockService)
	srv = Service{coreAPISrv: newCoreAPIService(docSrv)}
	h = handler{srv: srv}
	docSrv.On("CreateModel", mock.Anything, mock.Anything).Return(m, jobs.NewJobID(), nil)
	w, r = getHTTPReqAndResp(ctx, bytes.NewReader(marshall(t, data)))
	m.On("CalculateTransitionRulesFingerprint").Return(utils.RandomSlice(32), nil)
	h.CreateEntity(w, r)
	assert.Equal(t, w.Code, http.StatusAccepted)
	m.AssertExpectations(t)
	docSrv.AssertExpectations(t)
}

func TestHandler_UpdateEntity(t *testing.T) {
	getHTTPReqAndResp := func(ctx context.Context, b io.Reader) (*httptest.ResponseRecorder, *http.Request) {
		return httptest.NewRecorder(), httptest.NewRequest("PUT", "/entities/{document_id}", b).WithContext(ctx)
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
		h.UpdateEntity(w, r)
		assert.Equal(t, w.Code, http.StatusBadRequest)
		assert.Contains(t, w.Body.String(), coreapi.ErrInvalidDocumentID.Error())
	}

	// empty body
	id := hexutil.Encode(utils.RandomSlice(32))
	rctx.URLParams.Values[0] = id
	w, r := getHTTPReqAndResp(ctx, nil)
	h.UpdateEntity(w, r)
	assert.Equal(t, w.Code, http.StatusBadRequest)
	assert.Contains(t, w.Body.String(), "unexpected end of JSON input")

	// failed conversion
	data := map[string]interface{}{
		"data": entityData(),
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
	h.UpdateEntity(w, r)
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
	h.UpdateEntity(w, r)
	assert.Equal(t, w.Code, http.StatusBadRequest)
	assert.Contains(t, w.Body.String(), "failed to update model")
	docSrv.AssertExpectations(t)

	// failed response conversion
	m := new(testingdocuments.MockModel)
	m.On("GetData").Return(entity.Data{})
	m.On("Scheme").Return(entity.Scheme)
	m.On("GetAttributes").Return(nil)
	m.On("GetCollaborators", mock.Anything).Return(documents.CollaboratorsAccess{}, errors.New("failed to get collaborators"))
	docSrv = new(testingdocuments.MockService)
	srv = Service{coreAPISrv: newCoreAPIService(docSrv)}
	h = handler{srv: srv}
	docSrv.On("UpdateModel", mock.Anything, mock.Anything).Return(m, jobs.NewJobID(), nil)
	w, r = getHTTPReqAndResp(ctx, bytes.NewReader(d))
	h.UpdateEntity(w, r)
	assert.Equal(t, w.Code, http.StatusInternalServerError)
	assert.Contains(t, w.Body.String(), "failed to get collaborators")
	docSrv.AssertExpectations(t)
	m.AssertExpectations(t)

	// success
	m = new(testingdocuments.MockModel)
	m.On("GetCollaborators", mock.Anything).Return(documents.CollaboratorsAccess{}, nil).Once()
	m.On("GetData").Return(entity.Data{})
	m.On("Scheme").Return(entity.Scheme)
	m.On("ID").Return(utils.RandomSlice(32)).Once()
	m.On("CurrentVersion").Return(utils.RandomSlice(32)).Once()
	m.On("Author").Return(nil, errors.New("somerror"))
	m.On("Timestamp").Return(nil, errors.New("somerror"))
	m.On("NFTs").Return(nil)
	m.On("GetAttributes").Return(nil)
	m.On("CalculateTransitionRulesFingerprint").Return(utils.RandomSlice(32), nil)
	collab := testingidentity.GenerateRandomDID()
	m.On("IsDIDCollaborator", collab).Return(false, nil).Once()
	ctx = context.WithValue(ctx, config.AccountHeaderKey, collab.String())
	docSrv = new(testingdocuments.MockService)
	srv = Service{coreAPISrv: newCoreAPIService(docSrv)}
	h = handler{srv: srv}
	docSrv.On("UpdateModel", mock.Anything, mock.Anything).Return(m, jobs.NewJobID(), nil)
	w, r = getHTTPReqAndResp(ctx, bytes.NewReader(d))
	h.UpdateEntity(w, r)
	assert.Equal(t, w.Code, http.StatusAccepted)
	m.AssertExpectations(t)
	docSrv.AssertExpectations(t)
}

func TestHandler_GetEntity(t *testing.T) {
	getHTTPReqAndResp := func(ctx context.Context) (*httptest.ResponseRecorder, *http.Request) {
		return httptest.NewRecorder(), httptest.NewRequest("GET", "/entities/{document_id}", nil).WithContext(ctx)
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
		h.GetEntity(w, r)
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
	h.GetEntity(w, r)
	assert.Equal(t, w.Code, http.StatusNotFound)
	assert.Contains(t, w.Body.String(), coreapi.ErrDocumentNotFound.Error())
	docSrv.AssertExpectations(t)

	// failed doc response
	m := new(testingdocuments.MockModel)
	m.On("GetData").Return(entity.Data{})
	m.On("Scheme").Return(entity.Scheme)
	m.On("GetAttributes").Return(nil)
	m.On("GetCollaborators", mock.Anything).Return(documents.CollaboratorsAccess{}, errors.New("failed to get collaborators"))
	docSrv = new(testingdocuments.MockService)
	docSrv.On("GetCurrentVersion", id).Return(m, nil)
	h = handler{srv: Service{coreAPISrv: newCoreAPIService(docSrv)}}
	w, r = getHTTPReqAndResp(ctx)
	h.GetEntity(w, r)
	assert.Equal(t, w.Code, http.StatusInternalServerError)
	assert.Contains(t, w.Body.String(), "failed to get collaborators")
	docSrv.AssertExpectations(t)
	m.AssertExpectations(t)

	// success
	m = new(testingdocuments.MockModel)
	m.On("GetCollaborators", mock.Anything).Return(documents.CollaboratorsAccess{}, nil).Once()
	m.On("GetData").Return(entity.Data{})
	m.On("Scheme").Return(entity.Scheme)
	m.On("ID").Return(utils.RandomSlice(32)).Once()
	m.On("CurrentVersion").Return(utils.RandomSlice(32)).Once()
	m.On("Author").Return(nil, errors.New("somerror"))
	m.On("Timestamp").Return(nil, errors.New("somerror"))
	m.On("NFTs").Return(nil)
	m.On("GetAttributes").Return(nil)
	m.On("CalculateTransitionRulesFingerprint").Return(utils.RandomSlice(32), nil)
	collab := testingidentity.GenerateRandomDID()
	m.On("IsDIDCollaborator", collab).Return(false, nil).Once()
	ctx = context.WithValue(ctx, config.AccountHeaderKey, collab.String())
	docSrv = new(testingdocuments.MockService)
	docSrv.On("GetCurrentVersion", id).Return(m, nil)
	h = handler{srv: Service{coreAPISrv: newCoreAPIService(docSrv)}}
	w, r = getHTTPReqAndResp(ctx)
	h.GetEntity(w, r)
	assert.Equal(t, w.Code, http.StatusOK)
	docSrv.AssertExpectations(t)
	m.AssertExpectations(t)
}

func TestHandler_ShareEntity(t *testing.T) {
	getHTTPReqAndResp := func(ctx context.Context, b io.Reader) (*httptest.ResponseRecorder, *http.Request) {
		return httptest.NewRecorder(), httptest.NewRequest("GET", "/entities/{document_id}/share", b).WithContext(ctx)
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
		w, r := getHTTPReqAndResp(ctx, nil)
		h.ShareEntity(w, r)
		assert.Equal(t, w.Code, http.StatusBadRequest)
		assert.Contains(t, w.Body.String(), coreapi.ErrInvalidDocumentID.Error())
	}

	// empty body
	id := hexutil.Encode(utils.RandomSlice(32))
	rctx.URLParams.Values[0] = id
	w, r := getHTTPReqAndResp(ctx, nil)
	h.ShareEntity(w, r)
	assert.Equal(t, w.Code, http.StatusBadRequest)
	assert.Contains(t, w.Body.String(), "unexpected end of JSON input")

	// failed creation
	docSrv := new(testingdocuments.MockService)
	m := new(testingdocuments.MockModel)
	h.srv.coreAPISrv = newCoreAPIService(docSrv)
	did := testingidentity.GenerateRandomDID()
	did1 := testingidentity.GenerateRandomDID()
	ctx = context.WithValue(ctx, config.AccountHeaderKey, did.String())
	docSrv.On("CreateModel", ctx, mock.Anything).Return(m, jobs.NewJobID(), errors.New("failed create")).Once()
	req := ShareEntityRequest{TargetIdentity: did1}
	d, err := json.Marshal(req)
	assert.NoError(t, err)
	w, r = getHTTPReqAndResp(ctx, bytes.NewReader(d))
	h.ShareEntity(w, r)
	assert.Equal(t, w.Code, http.StatusBadRequest)
	assert.Contains(t, w.Body.String(), "failed create")

	// failed convert
	docSrv.On("CreateModel", ctx, mock.Anything).Return(m, jobs.NewJobID(), nil).Once()
	m.On("GetCollaborators", mock.Anything).Return(documents.CollaboratorsAccess{}, errors.New("failed to get collaborators")).Once()
	w, r = getHTTPReqAndResp(ctx, bytes.NewReader(d))
	h.ShareEntity(w, r)
	assert.Equal(t, w.Code, http.StatusInternalServerError)
	assert.Contains(t, w.Body.String(), "failed to get collaborators")

	// success
	id1 := byteutils.HexBytes(utils.RandomSlice(32))
	did2 := testingidentity.GenerateRandomDID()
	er := &entityrelationship.EntityRelationship{
		CoreDocument: &documents.CoreDocument{
			Document: coredocumentpb.CoreDocument{},
		},

		Data: entityrelationship.Data{
			TargetIdentity:   &did1,
			OwnerIdentity:    &did2,
			EntityIdentifier: id1,
		},
	}
	docSrv.On("CreateModel", ctx, mock.Anything).Return(er, jobs.NewJobID(), nil).Once()
	w, r = getHTTPReqAndResp(ctx, bytes.NewReader(d))
	h.ShareEntity(w, r)
	assert.Equal(t, w.Code, http.StatusAccepted)
	m.AssertExpectations(t)
	docSrv.AssertExpectations(t)
}

func TestHandler_RevokeEntity(t *testing.T) {
	getHTTPReqAndResp := func(ctx context.Context, b io.Reader) (*httptest.ResponseRecorder, *http.Request) {
		return httptest.NewRecorder(), httptest.NewRequest("GET", "/entities/{document_id}/revoke", b).WithContext(ctx)
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
		w, r := getHTTPReqAndResp(ctx, nil)
		h.RevokeEntity(w, r)
		assert.Equal(t, w.Code, http.StatusBadRequest)
		assert.Contains(t, w.Body.String(), coreapi.ErrInvalidDocumentID.Error())
	}

	// empty body
	id := hexutil.Encode(utils.RandomSlice(32))
	rctx.URLParams.Values[0] = id
	w, r := getHTTPReqAndResp(ctx, nil)
	h.RevokeEntity(w, r)
	assert.Equal(t, w.Code, http.StatusBadRequest)
	assert.Contains(t, w.Body.String(), "unexpected end of JSON input")

	// failed creation
	docSrv := new(testingdocuments.MockService)
	m := new(testingdocuments.MockModel)
	h.srv.coreAPISrv = newCoreAPIService(docSrv)
	did := testingidentity.GenerateRandomDID()
	did1 := testingidentity.GenerateRandomDID()
	ctx = context.WithValue(ctx, config.AccountHeaderKey, did.String())
	docSrv.On("UpdateModel", ctx, mock.Anything).Return(m, jobs.NewJobID(), errors.New("failed update")).Once()
	req := ShareEntityRequest{TargetIdentity: did1}
	d, err := json.Marshal(req)
	assert.NoError(t, err)
	w, r = getHTTPReqAndResp(ctx, bytes.NewReader(d))
	h.RevokeEntity(w, r)
	assert.Equal(t, w.Code, http.StatusBadRequest)
	assert.Contains(t, w.Body.String(), "failed update")

	// failed convert
	docSrv.On("UpdateModel", ctx, mock.Anything).Return(m, jobs.NewJobID(), nil).Once()
	m.On("GetCollaborators", mock.Anything).Return(documents.CollaboratorsAccess{}, errors.New("failed to get collaborators")).Once()
	w, r = getHTTPReqAndResp(ctx, bytes.NewReader(d))
	h.RevokeEntity(w, r)
	assert.Equal(t, w.Code, http.StatusInternalServerError)
	assert.Contains(t, w.Body.String(), "failed to get collaborators")

	// success
	id1 := byteutils.HexBytes(utils.RandomSlice(32))
	did2 := testingidentity.GenerateRandomDID()
	er := &entityrelationship.EntityRelationship{
		CoreDocument: &documents.CoreDocument{
			Document: coredocumentpb.CoreDocument{},
		},

		Data: entityrelationship.Data{
			TargetIdentity:   &did1,
			OwnerIdentity:    &did2,
			EntityIdentifier: id1,
		},
	}
	docSrv.On("UpdateModel", ctx, mock.Anything).Return(er, jobs.NewJobID(), nil).Once()
	w, r = getHTTPReqAndResp(ctx, bytes.NewReader(d))
	h.RevokeEntity(w, r)
	assert.Equal(t, w.Code, http.StatusAccepted)
	assert.Contains(t, w.Body.String(), "\"active\":false")
	m.AssertExpectations(t)
	docSrv.AssertExpectations(t)
}

func TestHandler_GetEntityThroughRelationship(t *testing.T) {
	getHTTPReqAndResp := func(ctx context.Context) (*httptest.ResponseRecorder, *http.Request) {
		return httptest.NewRecorder(), httptest.NewRequest("GET", "/relationships/{document_id}/entity", nil).WithContext(ctx)
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
		h.GetEntityThroughRelationship(w, r)
		assert.Equal(t, w.Code, http.StatusBadRequest)
		assert.Contains(t, w.Body.String(), coreapi.ErrInvalidDocumentID.Error())
	}

	// missing document
	id := utils.RandomSlice(32)
	rctx.URLParams.Values[0] = hexutil.Encode(id)
	eSrv := new(entity.MockService)
	eSrv.On("GetEntityByRelationship", mock.Anything, id).Return(nil, errors.New("failed")).Once()
	h = handler{srv: Service{entitySrv: eSrv}}
	w, r := getHTTPReqAndResp(ctx)
	h.GetEntityThroughRelationship(w, r)
	assert.Equal(t, w.Code, http.StatusNotFound)
	assert.Contains(t, w.Body.String(), coreapi.ErrDocumentNotFound.Error())

	// failed doc response
	m := new(testingdocuments.MockModel)
	m.On("GetData").Return(entity.Data{})
	m.On("Scheme").Return(entity.Scheme)
	m.On("GetAttributes").Return(nil)
	m.On("GetCollaborators", mock.Anything).Return(documents.CollaboratorsAccess{}, errors.New("failed to get collaborators")).Once()
	eSrv.On("GetEntityByRelationship", mock.Anything, id).Return(m, nil)
	w, r = getHTTPReqAndResp(ctx)
	h.GetEntityThroughRelationship(w, r)
	assert.Equal(t, w.Code, http.StatusInternalServerError)
	assert.Contains(t, w.Body.String(), "failed to get collaborators")

	// success
	m.On("GetCollaborators", mock.Anything).Return(documents.CollaboratorsAccess{}, nil).Once()
	m.On("ID").Return(utils.RandomSlice(32)).Once()
	m.On("CurrentVersion").Return(utils.RandomSlice(32)).Once()
	m.On("Author").Return(nil, errors.New("somerror"))
	m.On("Timestamp").Return(nil, errors.New("somerror"))
	m.On("NFTs").Return(nil)
	m.On("GetAttributes").Return(nil)
	collab := testingidentity.GenerateRandomDID()
	m.On("IsDIDCollaborator", collab).Return(false, nil).Once()
	m.On("CalculateTransitionRulesFingerprint").Return(utils.RandomSlice(32), nil)
	ctx = context.WithValue(ctx, config.AccountHeaderKey, collab.String())
	w, r = getHTTPReqAndResp(ctx)
	h.GetEntityThroughRelationship(w, r)
	assert.Equal(t, w.Code, http.StatusOK)
	eSrv.AssertExpectations(t)
	m.AssertExpectations(t)
}
