package userapi

import (
	"encoding/json"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/httpapi/coreapi"
	"github.com/centrifuge/go-centrifuge/jobs"
	"github.com/centrifuge/go-centrifuge/utils/httputils"
	"github.com/go-chi/render"
	logging "github.com/ipfs/go-log"
	"io/ioutil"
	"net/http"
)

type handler struct {
	srv           Service
	tokenRegistry documents.TokenRegistry
}

var log = logging.Logger("user-api")

// CreateTransferDetail creates a document.
// @summary Creates a new transfer detail extension on a document and anchors it.
// @description Creates a new transfer detail extension on a document and anchors it.
// @id create_transfer_detail
// @tags TransferDetail
// @accept json
// @param authorization header string true "Hex encoded centrifuge ID of the account for the intended API action"
// @param body body coreapi.CreateTransferDetailRequest true "Transfer Detail Create Request"
// @produce json
// @Failure 500 {object} httputils.HTTPError
// @Failure 400 {object} httputils.HTTPError
// @Failure 403 {object} httputils.HTTPError
// @success 201 {object} coreapi.TransferDetailResponse
// @router /v1/documents [post]
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

	payload, err := toTransferDetailCreatePayload(request)
	if err != nil {
		code = http.StatusBadRequest
		log.Error(err)
		return
	}
	model, err := h.srv.CreateTransferDetails(ctx, *payload)
	if err != nil {
		code = http.StatusBadRequest
		log.Error(err)
		return
	}

	header, err := coreapi.DeriveResponseHeader(h.tokenRegistry, model, jobs.NilJobID())
	if err != nil {
		code = http.StatusInternalServerError
		log.Error(err)
		return
	}

	resp := TransferDetailResponse{
		Header: &header,
		Data: payload.Data,
	}

	render.Status(r, http.StatusCreated)
	render.JSON(w, r, resp)
}


//Create
//Get
//GetVersion
//GetList
//GetListVersion
//Update
