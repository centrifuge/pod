package userapi

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/centrifuge/go-centrifuge/httpapi/coreapi"
	"github.com/centrifuge/go-centrifuge/utils/httputils"
	"github.com/go-chi/render"
)

// CreateInvoice creates an invoice document.
// @summary Creates a new invoice document and anchors it.
// @description Creates a new invoice document and anchors it.
// @id create_invoice
// @tags Invoice
// @accept json
// @param authorization header string true "Hex encoded centrifuge ID of the account for the intended API action"
// @param body body userapi.CreateInvoiceRequest true "Invoice Create Request"
// @produce json
// @Failure 500 {object} httputils.HTTPError
// @Failure 400 {object} httputils.HTTPError
// @Failure 403 {object} httputils.HTTPError
// @success 201 {object} coreapi.DocumentResponse
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

	resp, err := coreapi.GetDocumentResponse(m, h.tokenRegistry, j)
	if err != nil {
		code = http.StatusInternalServerError
		log.Error(err)
		return
	}

	render.Status(r, http.StatusCreated)
	render.JSON(w, r, resp)

}
