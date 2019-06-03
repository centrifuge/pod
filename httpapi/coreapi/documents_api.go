package coreapi

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/utils/httputils"
	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	logging "github.com/ipfs/go-log"
)

var log = logging.Logger("core_api")

type handler struct {
	srv           Service
	tokenRegistry documents.TokenRegistry
}

// CreateDocument creates a document.
// @summary Creates a new document and anchors it.
// @description Creates a new document and anchors it.
// @id create_document
// @tags Documents
// @accept json
// @param authorization header string true "centrifuge identity"
// @param body body coreapi.CreateDocumentRequest true "Document request"
// @produce json
// @Failure 500 {object} httputils.HTTPError
// @Failure 400 {object} httputils.HTTPError
// @Failure 403 {object} httputils.HTTPError
// @success 200 {object} coreapi.DocumentResponse
// @router /documents [post]
func (h handler) CreateDocument(w http.ResponseWriter, r *http.Request) {
	var err error
	var code int
	defer func() {
		if err == nil {
			return
		}

		render.Status(r, code)
		render.JSON(w, r, httputils.HTTPError{Message: err.Error()})
	}()

	ctx := r.Context()
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		code = http.StatusInternalServerError
		return
	}

	var request CreateDocumentRequest
	err = json.Unmarshal(data, &request)
	if err != nil {
		code = http.StatusBadRequest
		return
	}

	payload, err := toDocumentsCreatePayload(request)
	if err != nil {
		code = http.StatusBadRequest
		return
	}

	model, jobID, err := h.srv.CreateDocument(ctx, payload)
	if err != nil {
		code = http.StatusBadRequest
		return
	}

	docData := model.GetData()
	header, err := deriveResponseHeader(h.tokenRegistry, model, jobID)
	if err != nil {
		code = http.StatusInternalServerError
		return
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, DocumentResponse{Header: header, Data: docData})
}

// Register registers the core apis to the router.
func Register(r *chi.Mux, registry documents.TokenRegistry, docSrv documents.Service) {
	h := handler{srv: Service{docService: docSrv}, tokenRegistry: registry}
	r.Post("/documents", h.CreateDocument)
}
