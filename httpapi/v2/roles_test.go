// +build unit

package v2

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	coredocumentpb "github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/httpapi/coreapi"
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
	rctx.URLParams.Keys = make([]string, 2, 2)
	rctx.URLParams.Values = make([]string, 2, 2)
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
