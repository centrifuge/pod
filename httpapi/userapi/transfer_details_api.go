package userapi

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/extensions"
	"github.com/centrifuge/go-centrifuge/httpapi/coreapi"
	"github.com/centrifuge/go-centrifuge/jobs"
	"github.com/centrifuge/go-centrifuge/utils/httputils"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	logging "github.com/ipfs/go-log"
)

type handler struct {
	srv           Service
	tokenRegistry documents.TokenRegistry
}

var log = logging.Logger("user-api")

// CreateTransferDetail creates a transfer detail extension on a document.
// @summary Creates a new transfer detail extension on a document and anchors it.
// @description Creates a new transfer detail extension on a document and anchors it.
// @id create_transfer_detail
// @tags Transfer Details
// @accept json
// @param authorization header string true "Hex encoded centrifuge ID of the account for the intended API action"
// @param body body userapi.CreateTransferDetailRequest true "Transfer Detail Create Request"
// @param document_id path string true "Document Identifier"
// @produce json
// @Failure 500 {object} httputils.HTTPError
// @Failure 400 {object} httputils.HTTPError
// @Failure 403 {object} httputils.HTTPError
// @success 202 {object} userapi.TransferDetailResponse
// @router /v1/documents/{document_id}/transfer_details [post]
func (h handler) CreateTransferDetail(w http.ResponseWriter, r *http.Request) {
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

	var request CreateTransferDetailRequest
	err = json.Unmarshal(data, &request)
	if err != nil {
		code = http.StatusBadRequest
		log.Error(err)
		return
	}

	if request.Data.TransferID == "" {
		request.Data.TransferID = extensions.NewAttributeSetID()
	}
	request.DocumentID = chi.URLParam(r, coreapi.DocumentIDParam)

	payload, err := toTransferDetailCreatePayload(request)
	if err != nil {
		code = http.StatusBadRequest
		log.Error(err)
		return
	}

	model, jobID, err := h.srv.CreateTransferDetail(ctx, *payload)
	if err != nil {
		code = http.StatusBadRequest
		log.Error(err)
		return
	}

	header, err := coreapi.DeriveResponseHeader(h.tokenRegistry, model, jobID)
	if err != nil {
		code = http.StatusInternalServerError
		log.Error(err)
		return
	}

	resp := TransferDetailResponse{
		Header: header,
		Data:   payload.Data,
	}

	render.Status(r, http.StatusAccepted)
	render.JSON(w, r, resp)
}

// UpdateTransferDetail updates a transfer detail extension.
// @summary Updates a new transfer detail extension on a document and anchors it.
// @description Updates a new transfer detail extension on a document and anchors it.
// @id update_transfer_detail
// @tags Transfer Details
// @accept json
// @param authorization header string true "Hex encoded centrifuge ID of the account for the intended API action"
// @param body body userapi.UpdateTransferDetailRequest true "Transfer Detail Update Request"
// @param document_id path string true "Document Identifier"
// @param transfer_id path string true "Transfer Detail Identifier"
// @produce json
// @Failure 500 {object} httputils.HTTPError
// @Failure 400 {object} httputils.HTTPError
// @Failure 403 {object} httputils.HTTPError
// @success 202 {object} userapi.TransferDetailResponse
// @router /v1/documents/{document_id}/transfer_details/{transfer_id} [put]
func (h handler) UpdateTransferDetail(w http.ResponseWriter, r *http.Request) {
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

	var request UpdateTransferDetailRequest
	err = json.Unmarshal(data, &request)
	if err != nil {
		code = http.StatusBadRequest
		log.Error(err)
		return
	}

	request.TransferID = chi.URLParam(r, transferIDParam)
	request.DocumentID = chi.URLParam(r, coreapi.DocumentIDParam)
	payload, err := toTransferDetailUpdatePayload(request)
	if err != nil {
		code = http.StatusBadRequest
		log.Error(err)
		return
	}

	model, jobID, err := h.srv.UpdateTransferDetail(ctx, *payload)
	if err != nil {
		code = http.StatusBadRequest
		log.Error(err)
		return
	}

	header, err := coreapi.DeriveResponseHeader(h.tokenRegistry, model, jobID)
	if err != nil {
		code = http.StatusInternalServerError
		log.Error(err)
		return
	}

	resp := TransferDetailResponse{
		Header: header,
		Data:   payload.Data,
	}

	render.Status(r, http.StatusAccepted)
	render.JSON(w, r, resp)
}

// GetTransferDetail returns the latest version of transfer detail.
// @summary Returns the latest version of the transfer detail.
// @description Returns the latest version of the transfer detail.
// @id get_transfer_detail
// @tags Transfer Details
// @param authorization header string true "Hex encoded centrifuge ID of the account for the intended API action"
// @param document_id path string true "Document Identifier"
// @param transfer_id path string true "Transfer Detail Identifier"
// @produce json
// @Failure 403 {object} httputils.HTTPError
// @Failure 400 {object} httputils.HTTPError
// @Failure 404 {object} httputils.HTTPError
// @Failure 500 {object} httputils.HTTPError
// @success 200 {object} userapi.TransferDetailResponse
// @router /v1/documents/{document_id}/transfer_details/{transfer_id} [get]
func (h handler) GetTransferDetail(w http.ResponseWriter, r *http.Request) {
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

	transferID, err := hexutil.Decode(chi.URLParam(r, transferIDParam))
	if err != nil {
		code = http.StatusBadRequest
		log.Error(err)
		err = coreapi.ErrInvalidDocumentID
		return
	}

	td, model, err := h.srv.GetCurrentTransferDetail(r.Context(), docID, transferID)
	if err != nil {
		code = http.StatusNotFound
		log.Error(err)
		err = coreapi.ErrDocumentNotFound
		return
	}

	header, err := coreapi.DeriveResponseHeader(h.tokenRegistry, model, jobs.NilJobID())
	if err != nil {
		code = http.StatusInternalServerError
		log.Error(err)
		return
	}

	resp := TransferDetailResponse{
		Header: header,
		Data:   td.Data,
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, resp)
}

// GetTransferDetailList returns a list of the latest versions all transfer details on the document.
// @summary Returns a list of the latest versions of all transfer details on the document.
// @description Returns a list of the latest versions of all transfer details on the document.
// @id list_transfer_details
// @tags Transfer Details
// @param authorization header string true "Hex encoded centrifuge ID of the account for the intended API action"
// @param document_id path string true "Document Identifier"
// @produce json
// @Failure 403 {object} httputils.HTTPError
// @Failure 400 {object} httputils.HTTPError
// @Failure 404 {object} httputils.HTTPError
// @Failure 500 {object} httputils.HTTPError
// @success 200 {object} userapi.TransferDetailListResponse
// @router /v1/documents/{document_id}/transfer_details [get]
func (h handler) GetTransferDetailList(w http.ResponseWriter, r *http.Request) {
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

	td, model, err := h.srv.GetCurrentTransferDetailsList(r.Context(), docID)
	if err != nil {
		code = http.StatusNotFound
		log.Error(err)
		err = coreapi.ErrDocumentNotFound
		return
	}

	header, err := coreapi.DeriveResponseHeader(h.tokenRegistry, model, jobs.NilJobID())
	if err != nil {
		code = http.StatusInternalServerError
		log.Error(err)
		return
	}

	resp := TransferDetailListResponse{
		Header: header,
		Data:   td.Data,
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, resp)
}

//GetVersion
//GetListVersion
