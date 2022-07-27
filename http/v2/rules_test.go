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

	coredocumentpb "github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/http/coreapi"
	"github.com/centrifuge/go-centrifuge/pending"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/go-chi/chi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestHandler_AddTransitionRules(t *testing.T) {
	getHTTPReqAndResp := func(ctx context.Context, b io.Reader) (*httptest.ResponseRecorder, *http.Request) {
		return httptest.NewRecorder(), httptest.NewRequest("post", "/documents/{document_id}/transition_rules", b).WithContext(ctx)
	}

	// invalid doc id
	rctx := chi.NewRouteContext()
	rctx.URLParams.Keys = make([]string, 1)
	rctx.URLParams.Values = make([]string, 1)
	rctx.URLParams.Keys[0] = coreapi.DocumentIDParam
	rctx.URLParams.Values[0] = "some invalid id"
	ctx := context.WithValue(context.Background(), chi.RouteCtxKey, rctx)
	w, r := getHTTPReqAndResp(ctx, nil)
	h := handler{}
	h.AddTransitionRules(w, r)
	assert.Equal(t, w.Code, http.StatusBadRequest)
	assert.Contains(t, w.Body.String(), coreapi.ErrInvalidDocumentID.Error())

	type attrRule struct {
		KeyLabel string `json:"key_label"`
		RoleID   string `json:"role_id"`
	}

	var rule struct {
		AttributeRules []attrRule `json:"attribute_rules"`
	}

	// bad roleID
	rule.AttributeRules = make([]attrRule, 1)
	rule.AttributeRules[0].RoleID = ""
	rule.AttributeRules[0].KeyLabel = "test"
	d, err := json.Marshal(rule)
	assert.NoError(t, err)
	docID := utils.RandomSlice(32)
	rctx.URLParams.Values[0] = hexutil.Encode(docID)
	w, r = getHTTPReqAndResp(ctx, bytes.NewReader(d))
	h.AddTransitionRules(w, r)
	assert.Equal(t, w.Code, http.StatusBadRequest)
	assert.Contains(t, w.Body.String(), "empty hex string")

	// failed to add rule
	roleID := utils.RandomSlice(32)
	rule.AttributeRules[0].RoleID = hexutil.Encode(roleID)
	d, err = json.Marshal(rule)
	assert.NoError(t, err)
	psrv := new(pending.MockService)
	psrv.On("AddTransitionRules", mock.Anything, docID, mock.Anything).
		Return(nil, errors.New("failed to add rule")).Once()
	h.srv.pendingDocSrv = psrv
	w, r = getHTTPReqAndResp(ctx, bytes.NewReader(d))
	h.AddTransitionRules(w, r)
	assert.Equal(t, w.Code, http.StatusBadRequest)
	assert.Contains(t, w.Body.String(), "failed to add rule")

	// success )
	psrv.On("AddTransitionRules", mock.Anything, docID, mock.Anything).
		Return([]*coredocumentpb.TransitionRule{
			{
				RuleKey:   utils.RandomSlice(32),
				Roles:     [][]byte{roleID},
				MatchType: coredocumentpb.FieldMatchType_FIELD_MATCH_TYPE_PREFIX,
				Field:     utils.RandomSlice(10),
				Action:    coredocumentpb.TransitionAction_TRANSITION_ACTION_EDIT,
			},
		}, nil).Once()
	w, r = getHTTPReqAndResp(ctx, bytes.NewReader(d))
	h.AddTransitionRules(w, r)
	assert.Equal(t, w.Code, http.StatusOK)
	psrv.AssertExpectations(t)
}

func TestHandler_GetTransitionRule(t *testing.T) {
	getHTTPReqAndResp := func(ctx context.Context) (*httptest.ResponseRecorder, *http.Request) {
		return httptest.NewRecorder(), httptest.NewRequest(
			"Get", "/documents/{document_id}/transition_rules/{rule_id}", nil).WithContext(ctx)
	}

	// invalid doc id
	rctx := chi.NewRouteContext()
	rctx.URLParams.Keys = make([]string, 2, 2)
	rctx.URLParams.Values = make([]string, 2, 2)
	rctx.URLParams.Keys[0] = coreapi.DocumentIDParam
	rctx.URLParams.Keys[1] = RuleIDParam
	rctx.URLParams.Values[0] = "some invalid id"
	ctx := context.WithValue(context.Background(), chi.RouteCtxKey, rctx)
	w, r := getHTTPReqAndResp(ctx)
	h := handler{}
	h.GetTransitionRule(w, r)
	assert.Equal(t, w.Code, http.StatusBadRequest)
	assert.Contains(t, w.Body.String(), coreapi.ErrInvalidDocumentID.Error())

	// invalid rule ID
	docID := utils.RandomSlice(32)
	rctx.URLParams.Values[0] = hexutil.Encode(docID)
	rctx.URLParams.Values[1] = "some ruleID"
	w, r = getHTTPReqAndResp(ctx)
	h.GetTransitionRule(w, r)
	assert.Equal(t, w.Code, http.StatusBadRequest)
	assert.Contains(t, w.Body.String(), ErrInvalidRuleID.Error())

	// missing document or rule
	ruleID := utils.RandomSlice(32)
	rctx.URLParams.Values[1] = hexutil.Encode(ruleID)
	psrv := new(pending.MockService)
	psrv.On("GetTransitionRule", mock.Anything, docID, ruleID).Return(nil, errors.New("NotFound")).Once()
	h.srv.pendingDocSrv = psrv
	w, r = getHTTPReqAndResp(ctx)
	h.GetTransitionRule(w, r)
	assert.Equal(t, w.Code, http.StatusNotFound)
	assert.Contains(t, w.Body.String(), "NotFound")

	// success
	psrv.On("GetTransitionRule", mock.Anything, docID, ruleID).Return(
		new(coredocumentpb.TransitionRule), nil).Once()
	w, r = getHTTPReqAndResp(ctx)
	h.GetTransitionRule(w, r)
	assert.Equal(t, w.Code, http.StatusOK)
	psrv.AssertExpectations(t)
}

func TestHandler_DeleteTransitionRule(t *testing.T) {
	getHTTPReqAndResp := func(ctx context.Context) (*httptest.ResponseRecorder, *http.Request) {
		return httptest.NewRecorder(), httptest.NewRequest(
			"Delete", "/documents/{document_id}/transition_rules/{rule_id}", nil).WithContext(ctx)
	}

	// invalid doc id
	rctx := chi.NewRouteContext()
	rctx.URLParams.Keys = make([]string, 2, 2)
	rctx.URLParams.Values = make([]string, 2, 2)
	rctx.URLParams.Keys[0] = coreapi.DocumentIDParam
	rctx.URLParams.Keys[1] = RuleIDParam
	rctx.URLParams.Values[0] = "some invalid id"
	ctx := context.WithValue(context.Background(), chi.RouteCtxKey, rctx)
	w, r := getHTTPReqAndResp(ctx)
	h := handler{}
	h.DeleteTransitionRule(w, r)
	assert.Equal(t, w.Code, http.StatusBadRequest)
	assert.Contains(t, w.Body.String(), coreapi.ErrInvalidDocumentID.Error())

	// invalid rule ID
	docID := utils.RandomSlice(32)
	rctx.URLParams.Values[0] = hexutil.Encode(docID)
	rctx.URLParams.Values[1] = "some ruleID"
	w, r = getHTTPReqAndResp(ctx)
	h.DeleteTransitionRule(w, r)
	assert.Equal(t, w.Code, http.StatusBadRequest)
	assert.Contains(t, w.Body.String(), ErrInvalidRuleID.Error())

	// missing document or rule
	ruleID := utils.RandomSlice(32)
	rctx.URLParams.Values[1] = hexutil.Encode(ruleID)
	psrv := new(pending.MockService)
	psrv.On("DeleteTransitionRule", mock.Anything, docID, ruleID).Return(errors.New("NotFound")).Once()
	h.srv.pendingDocSrv = psrv
	w, r = getHTTPReqAndResp(ctx)
	h.DeleteTransitionRule(w, r)
	assert.Equal(t, w.Code, http.StatusNotFound)
	assert.Contains(t, w.Body.String(), "NotFound")

	// success
	psrv.On("DeleteTransitionRule", mock.Anything, docID, ruleID).Return(nil).Once()
	w, r = getHTTPReqAndResp(ctx)
	h.DeleteTransitionRule(w, r)
	assert.Equal(t, w.Code, http.StatusNoContent)
	psrv.AssertExpectations(t)
}
