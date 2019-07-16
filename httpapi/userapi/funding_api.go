package userapi

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/httpapi/coreapi"
	"github.com/centrifuge/go-centrifuge/jobs"
	"github.com/centrifuge/go-centrifuge/utils/httputils"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/go-chi/chi"
	"github.com/go-chi/render"
)

// ErrInvalidAgreementID is a sentinel error when the agreement ID is invalid.
const ErrInvalidAgreementID = errors.Error("Invalid funding agreement ID")

// CreateFundingAgreement creates a new funding agreement on the document associated with document_id.
// @summary Creates a new funding agreement on the document.
// @description Creates a new funding agreement on the document.
// @id create_funding_agreement
// @tags Funding Agreements
// @accept json
// @param authorization header string true "Hex encoded centrifuge ID of the account for the intended API action"
// @param document_id path string true "Document Identifier"
// @param body body userapi.FundingRequest true "Funding agreement Create Request"
// @produce json
// @Failure 500 {object} httputils.HTTPError
// @Failure 400 {object} httputils.HTTPError
// @Failure 403 {object} httputils.HTTPError
// @success 202 {object} userapi.FundingResponse
// @router /v1/documents/{document_id}/funding_agreements [post]
func (h handler) CreateFundingAgreement(w http.ResponseWriter, r *http.Request) {
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

	var request FundingRequest
	err = json.Unmarshal(data, &request)
	if err != nil {
		code = http.StatusBadRequest
		log.Error(err)
		return
	}

	m, j, err := h.srv.CreateFundingAgreement(ctx, docID, &request.Data)
	if err != nil {
		code = http.StatusBadRequest
		log.Error(err)
		return
	}

	resp, err := toFundingAgreementResponse(ctx, h.srv.fundingSrv, m, request.Data.AgreementID, h.tokenRegistry, j)
	if err != nil {
		code = http.StatusInternalServerError
		log.Error(err)
		return
	}

	render.Status(r, http.StatusAccepted)
	render.JSON(w, r, resp)
}

// GetFundingAgreements returns all the funding agreements in the document associated with document_id.
// @summary Returns all the funding agreements in the document associated with document_id.
// @description Returns all the funding agreements in the document associated with document_id.
// @id get_funding_agreements
// @tags Funding Agreements
// @accept json
// @param authorization header string true "Hex encoded centrifuge ID of the account for the intended API action"
// @param document_id path string true "Document Identifier"
// @produce json
// @Failure 500 {object} httputils.HTTPError
// @Failure 400 {object} httputils.HTTPError
// @Failure 404 {object} httputils.HTTPError
// @Failure 403 {object} httputils.HTTPError
// @success 200 {object} userapi.FundingListResponse
// @router /v1/documents/{document_id}/funding_agreements [get]
func (h handler) GetFundingAgreements(w http.ResponseWriter, r *http.Request) {
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
	m, err := h.srv.coreAPISrv.GetDocument(ctx, docID)
	if err != nil {
		code = http.StatusNotFound
		log.Error(err)
		return
	}

	resp, err := toFundingAgreementListResponse(ctx, h.srv.fundingSrv, m, h.tokenRegistry)
	if err != nil {
		code = http.StatusInternalServerError
		log.Error(err)
		return
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, resp)
}

// GetFundingAgreement returns the funding agreement associated with agreement_id in the document.
// @summary Returns the funding agreement associated with agreement_id in the document.
// @description Returns the funding agreement associated with agreement_id in the document.
// @id get_funding_agreement
// @tags Funding Agreements
// @accept json
// @param authorization header string true "Hex encoded centrifuge ID of the account for the intended API action"
// @param document_id path string true "Document Identifier"
// @param agreement_id path string true "Funding agreement Identifier"
// @produce json
// @Failure 500 {object} httputils.HTTPError
// @Failure 400 {object} httputils.HTTPError
// @Failure 404 {object} httputils.HTTPError
// @Failure 403 {object} httputils.HTTPError
// @success 200 {object} userapi.FundingResponse
// @router /v1/documents/{document_id}/funding_agreements/{agreement_id} [get]
func (h handler) GetFundingAgreement(w http.ResponseWriter, r *http.Request) {
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

	_, err = hexutil.Decode(chi.URLParam(r, agreementIDParam))
	if err != nil {
		code = http.StatusBadRequest
		log.Error(err)
		err = ErrInvalidAgreementID
		return
	}

	ctx := r.Context()
	m, err := h.srv.coreAPISrv.GetDocument(ctx, docID)
	if err != nil {
		code = http.StatusNotFound
		log.Error(err)
		return
	}

	resp, err := toFundingAgreementResponse(ctx, h.srv.fundingSrv, m, chi.URLParam(r, agreementIDParam), h.tokenRegistry, jobs.NilJobID())
	if err != nil {
		code = http.StatusInternalServerError
		log.Error(err)
		return
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, resp)
}

// UpdateFundingAgreement updates the funding agreement associated with agreement_id in the document.
// @summary Updates the funding agreement associated with agreement_id in the document.
// @description Updates the funding agreement associated with agreement_id in the document.
// @id update_funding_agreement
// @tags Funding Agreements
// @accept json
// @param authorization header string true "Hex encoded centrifuge ID of the account for the intended API action"
// @param document_id path string true "Document Identifier"
// @param agreement_id path string true "Funding agreement Identifier"
// @param body body userapi.FundingRequest true "Funding Agreement Update Request"
// @produce json
// @Failure 500 {object} httputils.HTTPError
// @Failure 400 {object} httputils.HTTPError
// @Failure 404 {object} httputils.HTTPError
// @Failure 403 {object} httputils.HTTPError
// @success 202 {object} userapi.FundingResponse
// @router /v1/documents/{document_id}/funding_agreements/{agreement_id} [PUT]
func (h handler) UpdateFundingAgreement(w http.ResponseWriter, r *http.Request) {
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

	fundingID, err := hexutil.Decode(chi.URLParam(r, agreementIDParam))
	if err != nil {
		code = http.StatusBadRequest
		log.Error(err)
		err = ErrInvalidAgreementID
		return
	}

	ctx := r.Context()
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		code = http.StatusInternalServerError
		log.Error(err)
		return
	}

	var request FundingRequest
	err = json.Unmarshal(data, &request)
	if err != nil {
		code = http.StatusBadRequest
		log.Error(err)
		return
	}

	m, jobID, err := h.srv.fundingSrv.UpdateFundingAgreement(ctx, docID, fundingID, &request.Data)
	if err != nil {
		code = http.StatusNotFound
		log.Error(err)
		return
	}

	resp, err := toFundingAgreementResponse(ctx, h.srv.fundingSrv, m, chi.URLParam(r, agreementIDParam), h.tokenRegistry, jobID)
	if err != nil {
		code = http.StatusInternalServerError
		log.Error(err)
		return
	}

	render.Status(r, http.StatusAccepted)
	render.JSON(w, r, resp)
}
