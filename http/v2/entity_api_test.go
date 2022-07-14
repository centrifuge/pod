//go:build unit
// +build unit

package v2

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/documents/entity"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/http/coreapi"
	testingdocuments "github.com/centrifuge/go-centrifuge/testingutils/documents"

	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/go-chi/chi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestHandler_GetEntityThroughRelationship(t *testing.T) {
	getHTTPReqAndResp := func(ctx context.Context) (*httptest.ResponseRecorder, *http.Request) {
		return httptest.NewRecorder(), httptest.NewRequest("GET", "/relationships/{document_id}/entity", nil).WithContext(ctx)
	}

	rctx := chi.NewRouteContext()
	rctx.URLParams.Keys = make([]string, 1)
	rctx.URLParams.Values = make([]string, 1)
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
	doc := new(testingdocuments.MockModel)
	doc.On("GetData").Return(entity.Data{})
	doc.On("Scheme").Return(entity.Scheme)
	doc.On("GetAttributes").Return(nil)
	doc.On("GetCollaborators", mock.Anything).Return(documents.CollaboratorsAccess{}, errors.New("failed to get collaborators")).Once()
	eSrv.On("GetEntityByRelationship", mock.Anything, id).Return(doc, nil)
	w, r = getHTTPReqAndResp(ctx)
	h.GetEntityThroughRelationship(w, r)
	assert.Equal(t, w.Code, http.StatusInternalServerError)
	assert.Contains(t, w.Body.String(), "failed to get collaborators")

	// success
	doc.On("GetCollaborators", mock.Anything).Return(documents.CollaboratorsAccess{}, nil).Once()
	doc.On("ID").Return(utils.RandomSlice(32)).Once()
	var prevID []byte = nil
	doc.On("PreviousVersion").Return(prevID).Once()
	doc.On("CurrentVersion").Return(utils.RandomSlice(32)).Once()
	doc.On("NextVersion").Return(utils.RandomSlice(32)).Once()
	doc.On("Author").Return(nil, errors.New("somerror"))
	doc.On("Timestamp").Return(nil, errors.New("somerror"))
	doc.On("NFTs").Return(nil)
	doc.On("GetAttributes").Return(nil)
	collab := testingidentity.GenerateRandomDID()
	doc.On("GetStatus").Return(documents.Pending).Once()
	doc.On("CalculateTransitionRulesFingerprint").Return(utils.RandomSlice(32), nil)
	ctx = context.WithValue(ctx, config.AccountHeaderKey, collab.String())
	w, r = getHTTPReqAndResp(ctx)
	h.GetEntityThroughRelationship(w, r)
	assert.Equal(t, w.Code, http.StatusOK)
	eSrv.AssertExpectations(t)
	doc.AssertExpectations(t)
}
