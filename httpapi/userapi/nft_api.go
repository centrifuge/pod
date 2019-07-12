package userapi

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/go-chi/render"

	"github.com/centrifuge/go-centrifuge/utils/httputils"
)

// MintInvoiceUnpaidNFT mints an NFT for an unpaid invoice document.
// @summary Mints an NFT for an unpaid invoice document.
// @description Mints an NFT for an unpaid invoice document.
// @id invoice_unpaid_nft
// @tags Invoices
// @accept json
// @param authorization header string true "Hex encoded centrifuge ID of the account for the intended API action"
// @param body body userapi.NFTMintInvoiceUnpaidRequest true "Invoice Unpaid NFT Mint Request"
// @produce json
// @Failure 500 {object} httputils.HTTPError
// @Failure 400 {object} httputils.HTTPError
// @Failure 403 {object} httputils.HTTPError
// @success 202 {object} userapi.NFTMintResponse
// @router /v1/invoices/{document_id}/mint/unpaid [post]
func (h handler) MintInvoiceUnpaidNFT(w http.ResponseWriter, r *http.Request) {
	var err error
	var code int
	defer httputils.RespondIfError(&code, &err, w, r)

	ctx := r.Context()
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		code = http.StatusInternalServerError
		log.Error(err)
		return
	}

	var request NFTMintInvoiceUnpaidRequest
	err = json.Unmarshal(data, &request)
	if err != nil {
		code = http.StatusBadRequest
		log.Error(err)
		return
	}

	m, err := h.srv.MintInvoiceUnpaidNFT(ctx, request)
	if err != nil {
		code = http.StatusBadRequest
		log.Error(err)
		return
	}

	render.Status(r, http.StatusAccepted)
	render.JSON(w, r, NFTMintResponse{
		Header: &ResponseHeader{JobId: m.JobID},
	})
}
