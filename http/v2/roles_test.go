//go:build unit
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

	coredocumentpb "github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/go-centrifuge/crypto"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/http/coreapi"

	"github.com/centrifuge/go-centrifuge/pending"

	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/centrifuge/go-centrifuge/utils/byteutils"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/go-chi/chi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestHandler_GetRole(t *testing.T) {
	getHTTPReqAndResp := func(ctx context.Context) (*httptest.ResponseRecorder, *http.Request) {
		return httptest.NewRecorder(), httptest.NewRequest("Get", "/documents/{document_id}/roles/{role_id}", nil).WithContext(ctx)
	}

	// invalid doc id
	rctx := chi.NewRouteContext()
	rctx.URLParams.Keys = make([]string, 2)
	rctx.URLParams.Values = make([]string, 2)
	rctx.URLParams.Keys[0] = coreapi.DocumentIDParam
	rctx.URLParams.Keys[1] = RoleIDParam
	rctx.URLParams.Values[0] = "some invalid id"
	ctx := context.WithValue(context.Background(), chi.RouteCtxKey, rctx)
	w, r := getHTTPReqAndResp(ctx)
	h := handler{}
	h.GetRole(w, r)
	assert.Equal(t, w.Code, http.StatusBadRequest)
	assert.Contains(t, w.Body.String(), coreapi.ErrInvalidDocumentID.Error())

	// invalid role ID
	docID := utils.RandomSlice(32)
	rctx.URLParams.Values[0] = hexutil.Encode(docID)
	rctx.URLParams.Values[1] = "some roleID"
	w, r = getHTTPReqAndResp(ctx)
	h.GetRole(w, r)
	assert.Equal(t, w.Code, http.StatusBadRequest)
	assert.Contains(t, w.Body.String(), ErrInvalidRoleID.Error())

	// missing document or role
	roleID := utils.RandomSlice(32)
	rctx.URLParams.Values[1] = hexutil.Encode(roleID)
	psrv := new(pending.MockService)
	psrv.On("GetRole", mock.Anything, docID, roleID).Return(nil, errors.New("NotFound")).Once()
	h.srv.pendingDocSrv = psrv
	w, r = getHTTPReqAndResp(ctx)
	h.GetRole(w, r)
	assert.Equal(t, w.Code, http.StatusNotFound)
	assert.Contains(t, w.Body.String(), "NotFound")

	// success
	collab := utils.RandomSlice(20)
	psrv.On("GetRole", mock.Anything, docID, roleID).Return(
		&coredocumentpb.Role{
			RoleKey:       roleID,
			Collaborators: [][]byte{collab},
		}, nil).Once()
	w, r = getHTTPReqAndResp(ctx)
	h.GetRole(w, r)
	assert.Equal(t, w.Code, http.StatusOK)
	res := w.Body.Bytes()
	var gr Role
	err := json.Unmarshal(res, &gr)
	assert.NoError(t, err)
	assert.Equal(t, Role{
		ID:            roleID,
		Collaborators: []byteutils.HexBytes{collab},
	}, gr)
	psrv.AssertExpectations(t)
}

func TestHandler_AddRole(t *testing.T) {
	getHTTPReqAndResp := func(ctx context.Context, b io.Reader) (*httptest.ResponseRecorder, *http.Request) {
		return httptest.NewRecorder(), httptest.NewRequest("post", "/documents/{document_id}/roles", b).WithContext(ctx)
	}

	// invalid doc id
	rctx := chi.NewRouteContext()
	rctx.URLParams.Keys = make([]string, 1, 1)
	rctx.URLParams.Values = make([]string, 1, 1)
	rctx.URLParams.Keys[0] = coreapi.DocumentIDParam
	rctx.URLParams.Values[0] = "some invalid id"
	ctx := context.WithValue(context.Background(), chi.RouteCtxKey, rctx)
	w, r := getHTTPReqAndResp(ctx, nil)
	h := handler{}
	h.AddRole(w, r)
	assert.Equal(t, w.Code, http.StatusBadRequest)
	assert.Contains(t, w.Body.String(), coreapi.ErrInvalidDocumentID.Error())

	var role struct {
		Key           string   `json:"key"`
		Collaborators []string `json:"collaborators"`
	}

	// bad collaborator address
	role.Key = "role label 1"
	role.Collaborators = []string{"invalid collaborator"}
	d, err := json.Marshal(role)
	assert.NoError(t, err)
	docID := utils.RandomSlice(32)
	rctx.URLParams.Values[0] = hexutil.Encode(docID)
	w, r = getHTTPReqAndResp(ctx, bytes.NewReader(d))
	h.AddRole(w, r)
	assert.Equal(t, w.Code, http.StatusBadRequest)
	assert.Contains(t, w.Body.String(), "malformed address provided")

	// failed to add role
	collab := testingidentity.GenerateRandomDID()
	role.Collaborators = []string{collab.String()}
	d, err = json.Marshal(role)
	assert.NoError(t, err)
	psrv := new(pending.MockService)
	psrv.On("AddRole", mock.Anything, docID, role.Key, []identity.DID{collab}).
		Return(nil, errors.New("failed to add role")).Once()
	h.srv.pendingDocSrv = psrv
	w, r = getHTTPReqAndResp(ctx, bytes.NewReader(d))
	h.AddRole(w, r)
	assert.Equal(t, w.Code, http.StatusBadRequest)
	assert.Contains(t, w.Body.String(), "failed to add role")

	// success
	id, err := crypto.Sha256Hash([]byte(role.Key))
	assert.NoError(t, err)
	psrv.On("AddRole", mock.Anything, docID, role.Key, []identity.DID{collab}).
		Return(&coredocumentpb.Role{
			RoleKey:       id,
			Collaborators: [][]byte{collab.ToAddress().Bytes()},
		}, nil).Once()
	w, r = getHTTPReqAndResp(ctx, bytes.NewReader(d))
	h.AddRole(w, r)
	assert.Equal(t, w.Code, http.StatusOK)
	res := w.Body.Bytes()
	var gr Role
	err = json.Unmarshal(res, &gr)
	assert.NoError(t, err)
	assert.Equal(t, Role{
		ID:            id,
		Collaborators: []byteutils.HexBytes{collab.ToAddress().Bytes()},
	}, gr)
	psrv.AssertExpectations(t)
}

func TestHandler_UpdateRole(t *testing.T) {
	getHTTPReqAndResp := func(ctx context.Context, b io.Reader) (*httptest.ResponseRecorder, *http.Request) {
		return httptest.NewRecorder(), httptest.NewRequest("patch", "/documents/{document_id}/roles/{role_id}", b).WithContext(ctx)
	}

	// invalid doc id
	rctx := chi.NewRouteContext()
	rctx.URLParams.Keys = make([]string, 2, 2)
	rctx.URLParams.Values = make([]string, 2, 2)
	rctx.URLParams.Keys[0] = coreapi.DocumentIDParam
	rctx.URLParams.Keys[1] = RoleIDParam
	rctx.URLParams.Values[0] = "some invalid id"
	ctx := context.WithValue(context.Background(), chi.RouteCtxKey, rctx)
	w, r := getHTTPReqAndResp(ctx, nil)
	h := handler{}
	h.UpdateRole(w, r)
	assert.Equal(t, w.Code, http.StatusBadRequest)
	assert.Contains(t, w.Body.String(), coreapi.ErrInvalidDocumentID.Error())

	// invalid role ID
	docID := utils.RandomSlice(32)
	rctx.URLParams.Values[0] = hexutil.Encode(docID)
	rctx.URLParams.Values[1] = "some roleID"
	w, r = getHTTPReqAndResp(ctx, nil)
	h.UpdateRole(w, r)
	assert.Equal(t, w.Code, http.StatusBadRequest)
	assert.Contains(t, w.Body.String(), ErrInvalidRoleID.Error())

	// bad address
	roleID := utils.RandomSlice(32)
	rctx.URLParams.Values[1] = hexutil.Encode(roleID)
	var role struct {
		Collaborators []string `json:"collaborators"`
	}
	role.Collaborators = []string{"invalid collaborator"}
	d, err := json.Marshal(role)
	assert.NoError(t, err)
	w, r = getHTTPReqAndResp(ctx, bytes.NewReader(d))
	h.UpdateRole(w, r)
	assert.Equal(t, w.Code, http.StatusBadRequest)
	assert.Contains(t, w.Body.String(), "malformed address provided")

	// missing document or role
	collab := testingidentity.GenerateRandomDID()
	role.Collaborators[0] = collab.String()
	d, err = json.Marshal(role)
	assert.NoError(t, err)
	w, r = getHTTPReqAndResp(ctx, bytes.NewReader(d))
	psrv := new(pending.MockService)
	psrv.On("UpdateRole", mock.Anything, docID, roleID, []identity.DID{collab}).Return(nil, errors.New("NotFound")).Once()
	h.srv.pendingDocSrv = psrv
	h.UpdateRole(w, r)
	assert.Equal(t, w.Code, http.StatusNotFound)
	assert.Contains(t, w.Body.String(), "NotFound")

	// success
	psrv.On("UpdateRole", mock.Anything, docID, roleID, []identity.DID{collab}).Return(&coredocumentpb.Role{}, nil).Once()
	w, r = getHTTPReqAndResp(ctx, bytes.NewReader(d))
	h.UpdateRole(w, r)
	assert.Equal(t, w.Code, http.StatusOK)
	psrv.AssertExpectations(t)
}
