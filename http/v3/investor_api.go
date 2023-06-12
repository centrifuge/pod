package v3

import (
	"net/http"

	"github.com/centrifuge/pod/errors"
	"github.com/centrifuge/pod/http/coreapi"
	"github.com/centrifuge/pod/utils/httputils"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/go-chi/render"
)

const (
	ErrInvalidAssetID = errors.Error("invalid asset ID")
)

// GetAsset retrieves an asset (document) that belongs to an issuer.
// @summary retrieves an asset (document) that belongs to an issuer.
// @description retrieves an asset (document) that belongs to an issuer.
// @id get_asset
// @tags Investor
// @accept json
// @param authorization header string true "Bearer <JW3T token>"
// @param asset_id path string true "Hex encoded asset (document) identifier"
// @param loan_id path string true "Loan ID"
// @param pool_id path string true "Pool ID"
// @produce json
// @Failure 403 {object} httputils.HTTPError
// @Failure 500 {object} httputils.HTTPError
// @Failure 400 {object} httputils.HTTPError
// @success 202 {object} coreapi.DocumentResponse
// @router /v3/investor/assets [get]
func (h *handler) GetAsset(w http.ResponseWriter, r *http.Request) {
	var err error
	var code int
	defer httputils.RespondIfError(&code, &err, w, r)

	assetID, err := hexutil.Decode(r.URL.Query().Get(coreapi.AssetIDQueryParam))

	if err != nil {
		code = http.StatusBadRequest
		err = ErrInvalidAssetID
		h.log.Error(err)
		return
	}

	doc, err := h.srv.GetDocument(r.Context(), assetID)

	if err != nil {
		code = http.StatusNotFound
		h.log.Error(err)
		err = coreapi.ErrDocumentNotFound
		return
	}

	res, err := coreapi.GetDocumentResponse(doc, "")

	if err != nil {
		code = http.StatusInternalServerError
		h.log.Error(err)
		return
	}

	res.Header.Status = string(doc.GetStatus())

	render.Status(r, http.StatusOK)
	render.JSON(w, r, res)
}
