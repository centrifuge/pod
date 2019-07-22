package userapi

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/centrifuge/go-centrifuge/httpapi/coreapi"
	"github.com/centrifuge/go-centrifuge/jobs"
	"github.com/centrifuge/go-centrifuge/utils/httputils"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/go-chi/chi"
	"github.com/go-chi/render"
)

// CreateInvoice creates an invoice document.
// @summary Creates a new invoice document and anchors it.
// @description Creates a new invoice document and anchors it.
// @id create_invoice
// @tags Invoices
// @accept json
// @param authorization header string true "Hex encoded centrifuge ID of the account for the intended API action"
// @param body body userapi.CreateInvoiceRequest true "Invoice Create Request"
// @produce json
// @Failure 500 {object} httputils.HTTPError
// @Failure 400 {object} httputils.HTTPError
// @Failure 403 {object} httputils.HTTPError
// @success 202 {object} userapi.InvoiceResponse
// @router /v1/invoices [post]
func (h handler) CreateInvoice(w http.ResponseWriter, r *http.Request) {
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

	var request CreateInvoiceRequest
	err = json.Unmarshal(data, &request)
	if err != nil {
		code = http.StatusBadRequest
		log.Error(err)
		return
	}

	m, j, err := h.srv.CreateInvoice(ctx, request)
	if err != nil {
		code = http.StatusBadRequest
		log.Error(err)
		return
	}

	resp, err := toInvoiceResponse(m, h.tokenRegistry, j)
	if err != nil {
		code = http.StatusInternalServerError
		log.Error(err)
		return
	}

	render.Status(r, http.StatusAccepted)
	render.JSON(w, r, resp)
}

// UpdateInvoice updates an existing Invoice.
// @summary Updates an existing Invoice and anchors it.
// @description Updates an existing invoice and anchors it.
// @id update_invoice
// @tags Invoices
// @accept json
// @param authorization header string true "Hex encoded centrifuge ID of the account for the intended API action"
// @param document_id path string true "Document Identifier"
// @param body body userapi.CreateInvoiceRequest true "Invoice Update request"
// @produce json
// @Failure 500 {object} httputils.HTTPError
// @Failure 400 {object} httputils.HTTPError
// @Failure 403 {object} httputils.HTTPError
// @success 202 {object} userapi.InvoiceResponse
// @router /v1/invoices/{document_id} [put]
func (h handler) UpdateInvoice(w http.ResponseWriter, r *http.Request) {
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
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		code = http.StatusInternalServerError
		log.Error(err)
		return
	}

	var request CreateInvoiceRequest
	err = json.Unmarshal(data, &request)
	if err != nil {
		code = http.StatusBadRequest
		log.Error(err)
		return
	}

	model, jobID, err := h.srv.UpdateInvoice(ctx, docID, request)
	if err != nil {
		code = http.StatusBadRequest
		log.Error(err)
		return
	}

	resp, err := toInvoiceResponse(model, h.tokenRegistry, jobID)
	if err != nil {
		code = http.StatusInternalServerError
		log.Error(err)
		return
	}

	render.Status(r, http.StatusAccepted)
	render.JSON(w, r, resp)
}

// GetInvoice returns the latest version of the Invoice.
// @summary Returns the latest version of the Invoice.
// @description Returns the latest version of the Invoice.
// @id get_invoice
// @tags Invoices
// @param authorization header string true "Hex encoded centrifuge ID of the account for the intended API action"
// @param document_id path string true "Document Identifier"
// @produce json
// @Failure 403 {object} httputils.HTTPError
// @Failure 400 {object} httputils.HTTPError
// @Failure 404 {object} httputils.HTTPError
// @Failure 500 {object} httputils.HTTPError
// @success 200 {object} userapi.InvoiceResponse
// @router /v1/invoices/{document_id} [get]
func (h handler) GetInvoice(w http.ResponseWriter, r *http.Request) {
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

	model, err := h.srv.GetInvoice(r.Context(), docID)
	if err != nil {
		code = http.StatusNotFound
		log.Error(err)
		err = coreapi.ErrDocumentNotFound
		return
	}

	resp, err := toInvoiceResponse(model, h.tokenRegistry, jobs.NilJobID())
	if err != nil {
		code = http.StatusInternalServerError
		log.Error(err)
		return
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, resp)
}

// GetInvoiceVersion returns the specific version of an Invoice.
// @summary Returns the specific version of an Invoice.
// @description Returns the specific version of an Invoice.
// @id get_invoice_version
// @tags Invoices
// @param authorization header string true "Hex encoded centrifuge ID of the account for the intended API action"
// @param document_id path string true "Document Identifier"
// @param version_id path string true "Document Version Identifier"
// @produce json
// @Failure 403 {object} httputils.HTTPError
// @Failure 400 {object} httputils.HTTPError
// @Failure 404 {object} httputils.HTTPError
// @Failure 500 {object} httputils.HTTPError
// @success 200 {object} userapi.InvoiceResponse
// @router /v1/invoices/{document_id}/versions/{version_id} [get]
func (h handler) GetInvoiceVersion(w http.ResponseWriter, r *http.Request) {
	var err error
	var code int
	defer httputils.RespondIfError(&code, &err, w, r)

	ids := make([][]byte, 2, 2)
	for i, idStr := range []string{chi.URLParam(r, coreapi.DocumentIDParam), chi.URLParam(r, coreapi.VersionIDParam)} {
		var id []byte
		id, err = hexutil.Decode(idStr)
		if err != nil {
			code = http.StatusBadRequest
			log.Error(err)
			err = coreapi.ErrInvalidDocumentID
			return
		}

		ids[i] = id
	}

	model, err := h.srv.GetInvoiceVersion(r.Context(), ids[0], ids[1])
	if err != nil {
		code = http.StatusNotFound
		log.Error(err)
		err = coreapi.ErrDocumentNotFound
		return
	}

	resp, err := toInvoiceResponse(model, h.tokenRegistry, jobs.NilJobID())
	if err != nil {
		code = http.StatusInternalServerError
		log.Error(err)
		return
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, resp)
}

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

	docID, err := hexutil.Decode(chi.URLParam(r, coreapi.DocumentIDParam))
	if err != nil {
		code = http.StatusBadRequest
		log.Error(err)
		err = coreapi.ErrInvalidDocumentID
		return
	}

	ctx := r.Context()
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		code = http.StatusInternalServerError
		log.Error(err)
		return
	}

	var req NFTMintInvoiceUnpaidRequest
	err = json.Unmarshal(data, &req)
	if err != nil {
		code = http.StatusBadRequest
		log.Error(err)
		return
	}

	m, err := h.srv.MintInvoiceUnpaidNFT(ctx, docID, req.DepositAddress)
	if err != nil {
		code = http.StatusBadRequest
		log.Error(err)
		return
	}

	render.Status(r, http.StatusAccepted)
	render.JSON(w, r, NFTMintResponse{
		Header: &ResponseHeader{JobID: m.JobID},
	})
}
