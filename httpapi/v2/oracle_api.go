package v2

import (
	"net/http"

	"github.com/centrifuge/go-centrifuge/httpapi/coreapi"
	"github.com/centrifuge/go-centrifuge/oracle"
	"github.com/centrifuge/go-centrifuge/utils/httputils"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/go-chi/chi"
	"github.com/go-chi/render"
)

// PushAttributeToOracle pushes a given attribute value to oracle.
// @summary Pushes a given attribute value to oracle.
// @description Pushes a given attribute value to oracle.
// @id push_attribute_oracle
// @tags NFTs
// @accept json
// @param authorization header string true "Hex encoded centrifuge ID of the account for the intended API action"
// @param body body oracle.PushAttributeToOracleRequest true "Push Attribute to Oracle Request"
// @param document_id path string true "Document Identifier"
// @produce json
// @Failure 400 {object} httputils.HTTPError
// @Failure 403 {object} httputils.HTTPError
// @success 200 {object} oracle.PushToOracleResponse
// @router /v2/documents/{document_id}/push_to_oracle [post]
func (h handler) PushAttributeToOracle(w http.ResponseWriter, r *http.Request) {
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

	var req oracle.PushAttributeToOracleRequest
	err = unmarshalBody(r, &req)
	if err != nil {
		code = http.StatusBadRequest
		log.Error(err)
		return
	}

	resp, err := h.srv.PushAttributeToOracle(r.Context(), docID, req)
	if err != nil {
		code = http.StatusBadRequest
		log.Error(err)
		return
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, resp)
}
