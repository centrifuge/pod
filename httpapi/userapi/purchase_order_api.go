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

// CreatePurchaseOrder creates a new purchase order.
// @summary Creates a new purchase order and anchors it.
// @description Creates a new purchase order and anchors it.
// @id create_purchase_order
// @tags Purchase Orders
// @accept json
// @param authorization header string true "Hex encoded centrifuge ID of the account for the intended API action"
// @param body body userapi.CreatePurchaseOrderRequest true "Purchase Order Create request"
// @produce json
// @Failure 500 {object} httputils.HTTPError
// @Failure 400 {object} httputils.HTTPError
// @Failure 403 {object} httputils.HTTPError
// @success 201 {object} userapi.PurchaseOrderResponse
// @router  /v1/purchase_orders [post]
func (h handler) CreatePurchaseOrder(w http.ResponseWriter, r *http.Request) {
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

	var request CreatePurchaseOrderRequest
	err = json.Unmarshal(data, &request)
	if err != nil {
		code = http.StatusBadRequest
		log.Error(err)
		return
	}

	model, jobID, err := h.srv.CreatePurchaseOrder(ctx, request)
	if err != nil {
		code = http.StatusBadRequest
		log.Error(err)
		return
	}

	resp, err := toPurchaseOrderResponse(model, h.tokenRegistry, jobID)
	if err != nil {
		code = http.StatusInternalServerError
		log.Error(err)
		return
	}

	render.Status(r, http.StatusCreated)
	render.JSON(w, r, resp)
}

// GetPurchaseOrder returns the latest version of the PurchaseOrder.
// @summary Returns the latest version of the PurchaseOrder.
// @description Returns the latest version of the PurchaseOrder.
// @id get_purchase_order
// @tags Purchase Orders
// @param authorization header string true "Hex encoded centrifuge ID of the account for the intended API action"
// @param document_id path string true "Document Identifier"
// @produce json
// @Failure 403 {object} httputils.HTTPError
// @Failure 400 {object} httputils.HTTPError
// @Failure 404 {object} httputils.HTTPError
// @Failure 500 {object} httputils.HTTPError
// @success 200 {object} userapi.PurchaseOrderResponse
// @router /v1/purchase_orders/{document_id} [get]
func (h handler) GetPurchaseOrder(w http.ResponseWriter, r *http.Request) {
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

	model, err := h.srv.GetPurchaseOrder(r.Context(), docID)
	if err != nil {
		code = http.StatusNotFound
		log.Error(err)
		err = coreapi.ErrDocumentNotFound
		return
	}

	resp, err := toPurchaseOrderResponse(model, h.tokenRegistry, jobs.NilJobID())
	if err != nil {
		code = http.StatusInternalServerError
		log.Error(err)
		return
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, resp)
}

// UpdatePurchaseOrder updates an existing PurchaseOrder.
// @summary Updates an existing PurchaseOrder and anchors it.
// @description Updates an existing PurchaseOrder and anchors it.
// @id update_purchase_order
// @tags Purchase Orders
// @accept json
// @param authorization header string true "Hex encoded centrifuge ID of the account for the intended API action"
// @param document_id path string true "Document Identifier"
// @param body body userapi.CreatePurchaseOrderRequest true "PurchaseOrder Update request"
// @produce json
// @Failure 500 {object} httputils.HTTPError
// @Failure 400 {object} httputils.HTTPError
// @Failure 403 {object} httputils.HTTPError
// @success 201 {object} userapi.PurchaseOrderResponse
// @router /v1/purchase_orders/{document_id} [put]
func (h handler) UpdatePurchaseOrder(w http.ResponseWriter, r *http.Request) {
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

	var request CreatePurchaseOrderRequest
	err = json.Unmarshal(data, &request)
	if err != nil {
		code = http.StatusBadRequest
		log.Error(err)
		return
	}

	model, jobID, err := h.srv.UpdatePurchaseOrder(ctx, docID, request)
	if err != nil {
		code = http.StatusBadRequest
		log.Error(err)
		return
	}

	resp, err := toPurchaseOrderResponse(model, h.tokenRegistry, jobID)
	if err != nil {
		code = http.StatusInternalServerError
		log.Error(err)
		return
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, resp)
}

// GetPurchaseOrderVersion returns the specific version of a PurchaseOrder.
// @summary Returns the specific version of a PurchaseOrder.
// @description Returns the specific version of a PurchaseOrder.
// @id get_purchase_order_version
// @tags Purchase Orders
// @param authorization header string true "Hex encoded centrifuge ID of the account for the intended API action"
// @param document_id path string true "Document Identifier"
// @param version_id path string true "Document Version Identifier"
// @produce json
// @Failure 403 {object} httputils.HTTPError
// @Failure 400 {object} httputils.HTTPError
// @Failure 404 {object} httputils.HTTPError
// @Failure 500 {object} httputils.HTTPError
// @success 200 {object} userapi.PurchaseOrderResponse
// @router /v1/purchase_orders/{document_id}/versions/{version_id} [get]
func (h handler) GetPurchaseOrderVersion(w http.ResponseWriter, r *http.Request) {
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

	model, err := h.srv.GetPurchaseOrderVersion(r.Context(), ids[0], ids[1])
	if err != nil {
		code = http.StatusNotFound
		log.Error(err)
		err = coreapi.ErrDocumentNotFound
		return
	}

	resp, err := toPurchaseOrderResponse(model, h.tokenRegistry, jobs.NilJobID())
	if err != nil {
		code = http.StatusInternalServerError
		log.Error(err)
		return
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, resp)
}
