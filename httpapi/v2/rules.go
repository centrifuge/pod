package v2

import (
	"net/http"

	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/httpapi/coreapi"
	"github.com/centrifuge/go-centrifuge/pending"
	"github.com/centrifuge/go-centrifuge/utils/byteutils"
	"github.com/centrifuge/go-centrifuge/utils/httputils"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/go-chi/chi"
	"github.com/go-chi/render"
)

// RuleIDParam is the key for ruleID in the API path.
const RuleIDParam = "rule_id"

// ErrInvalidRuleID for invalid ruleID in the api path.
const ErrInvalidRuleID = errors.Error("Invalid Transition Rule ID")

// TransitionRule holds the ruleID, roles, and fields in hex format
type TransitionRule struct {
	RuleID byteutils.HexBytes   `json:"rule_id" swaggertype:"primitive,string"`
	Roles  []byteutils.HexBytes `json:"roles" swaggertype:"array,string"`
	Field  byteutils.HexBytes   `json:"field" swaggertype:"primitive,string"`
	Action string               `json:"action"`
}

// TransitionRules holds the list of transition rule.
type TransitionRules struct {
	Rules []TransitionRule `json:"rules"`
}

// AddTransitionRules adds a new transition rules to the document.
// @summary Adds a transition new rules to the document.
// @description Adds a new transition rules to the document.
// @id add_transition_rule
// @tags Documents
// @param authorization header string true "Hex encoded centrifuge ID of the account for the intended API action"
// @param document_id path string true "Document Identifier"
// @param body body pending.AddTransitionRules true "Add Transition rules Request"
// @produce json
// @Failure 403 {object} httputils.HTTPError
// @Failure 400 {object} httputils.HTTPError
// @Failure 404 {object} httputils.HTTPError
// @success 200 {object} v2.TransitionRules
// @router /v2/documents/{document_id}/transition_rules [post]
func (h handler) AddTransitionRules(w http.ResponseWriter, r *http.Request) {
	var err error
	var code int
	defer httputils.RespondIfError(&code, &err, w, r)

	docID, err := hexutil.Decode(chi.URLParam(r, coreapi.DocumentIDParam))
	if err != nil {
		code = http.StatusBadRequest
		log.Error(err)
		err = coreapi.ErrInvalidDocumentID
		return
	}

	ctx := r.Context()
	var addRules pending.AddTransitionRules
	err = unmarshalBody(r, &addRules)
	if err != nil {
		code = http.StatusBadRequest
		log.Error(err)
		return
	}

	rules, err := h.srv.AddTransitionRules(ctx, docID, addRules)
	if err != nil {
		code = http.StatusBadRequest
		log.Error(err)
		return
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, toClientRules(rules))
}

// GetTransitionRule returns the rule associated with the ruleID in the document
// @summary Returns the rule associated with the ruleID in the document.
// @description Returns the rule associated with the ruleID in the document.
// @id get_transition_rule
// @tags Documents
// @param authorization header string true "Hex encoded centrifuge ID of the account for the intended API action"
// @param document_id path string true "Document Identifier"
// @param rule_id path string true "Transition rule ID"
// @produce json
// @Failure 403 {object} httputils.HTTPError
// @Failure 400 {object} httputils.HTTPError
// @Failure 404 {object} httputils.HTTPError
// @success 200 {object} v2.TransitionRule
// @router /v2/documents/{document_id}/transition_rules/{rule_id} [get]
func (h handler) GetTransitionRule(w http.ResponseWriter, r *http.Request) {
	var err error
	var code int
	defer httputils.RespondIfError(&code, &err, w, r)

	docID, err := hexutil.Decode(chi.URLParam(r, coreapi.DocumentIDParam))
	if err != nil {
		code = http.StatusBadRequest
		log.Error(err)
		err = coreapi.ErrInvalidDocumentID
		return
	}

	ruleID, err := hexutil.Decode(chi.URLParam(r, RuleIDParam))
	if err != nil {
		code = http.StatusBadRequest
		log.Error(err)
		err = ErrInvalidRuleID
		return
	}

	rule, err := h.srv.GetTransitionRule(r.Context(), docID, ruleID)
	if err != nil {
		code = http.StatusNotFound
		log.Error(err)
		return
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, toClientRule(rule))
}
