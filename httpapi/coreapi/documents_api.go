package coreapi

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/jobs"
	"github.com/centrifuge/go-centrifuge/utils/httputils"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	logging "github.com/ipfs/go-log"
)

const (
	// ErrInvalidDocumentID is a sentinel error for invalid document identifiers.
	ErrInvalidDocumentID = errors.Error("invalid document identifier")

	// ErrDocumentNotFound is a sentinel error for missing documents.
	ErrDocumentNotFound = errors.Error("document not found")
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
// @param body body coreapi.CreateDocumentRequest true "Document Create request"
// @produce json
// @Failure 500 {object} httputils.HTTPError
// @Failure 400 {object} httputils.HTTPError
// @Failure 403 {object} httputils.HTTPError
// @success 200 {object} coreapi.DocumentResponse
// @router /documents [post]
func (h handler) CreateDocument(w http.ResponseWriter, r *http.Request) {
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

	var request CreateDocumentRequest
	err = json.Unmarshal(data, &request)
	if err != nil {
		code = http.StatusBadRequest
		log.Error(err)
		return
	}

	payload, err := toDocumentsCreatePayload(request)
	if err != nil {
		code = http.StatusBadRequest
		log.Error(err)
		return
	}

	model, jobID, err := h.srv.CreateDocument(ctx, payload)
	if err != nil {
		code = http.StatusBadRequest
		log.Error(err)
		return
	}

	resp, err := getDocumentResponse(model, h.tokenRegistry, jobID)
	if err != nil {
		code = http.StatusInternalServerError
		log.Error(err)
		return
	}

	render.Status(r, http.StatusCreated)
	render.JSON(w, r, resp)
}

// UpdateDocument updates an existing document.
// @summary Updates an existing document and anchors it.
// @description Updates an existing document and anchors it.
// @id update_document
// @tags Documents
// @accept json
// @param authorization header string true "centrifuge identity"
// @param body body coreapi.UpdateDocumentRequest true "Document Update request"
// @produce json
// @Failure 500 {object} httputils.HTTPError
// @Failure 400 {object} httputils.HTTPError
// @Failure 403 {object} httputils.HTTPError
// @success 200 {object} coreapi.DocumentResponse
// @router /documents [put]
func (h handler) UpdateDocument(w http.ResponseWriter, r *http.Request) {
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

	var request UpdateDocumentRequest
	err = json.Unmarshal(data, &request)
	if err != nil {
		code = http.StatusBadRequest
		log.Error(err)
		return
	}

	payload, err := toDocumentsUpdatePayload(request)
	if err != nil {
		code = http.StatusBadRequest
		log.Error(err)
		return
	}

	model, jobID, err := h.srv.UpdateDocument(ctx, payload)
	if err != nil {
		code = http.StatusBadRequest
		log.Error(err)
		return
	}

	resp, err := getDocumentResponse(model, h.tokenRegistry, jobID)
	if err != nil {
		code = http.StatusInternalServerError
		log.Error(err)
		return
	}

	render.Status(r, http.StatusCreated)
	render.JSON(w, r, resp)
}

// GetDocument returns the latest version of the document.
// @summary Returns the latest version of the document.
// @description Returns the latest version of the document.
// @id get_document
// @tags Documents
// @param authorization header string true "centrifuge identity"
// @param document_id path string true "Document Identifier"
// @produce json
// @Failure 400 {object} httputils.HTTPError
// @Failure 404 {object} httputils.HTTPError
// @Failure 500 {object} httputils.HTTPError
// @success 200 {object} coreapi.DocumentResponse
// @router /documents/{document_id} [get]
func (h handler) GetDocument(w http.ResponseWriter, r *http.Request) {
	var err error
	var code int
	defer httputils.RespondIfError(&code, &err, w, r)

	docID, err := hexutil.Decode(chi.URLParam(r, "document_id"))
	if err != nil {
		code = http.StatusBadRequest
		log.Error(err)
		err = ErrInvalidDocumentID
		return
	}

	model, err := h.srv.GetDocument(r.Context(), docID)
	if err != nil {
		code = http.StatusNotFound
		log.Error(err)
		err = ErrDocumentNotFound
		return
	}

	resp, err := getDocumentResponse(model, h.tokenRegistry, jobs.NilJobID())
	if err != nil {
		code = http.StatusInternalServerError
		log.Error(err)
		return
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, resp)
}

// GetDocumentVersion returns the specific version of the document.
// @summary Returns the specific version of the document.
// @description Returns the specific version of the document.
// @id get_document_version
// @tags Documents
// @param authorization header string true "centrifuge identity"
// @param document_id path string true "Document Identifier"
// @param version_id path string true "Document Version Identifier"
// @produce json
// @Failure 400 {object} httputils.HTTPError
// @Failure 404 {object} httputils.HTTPError
// @Failure 500 {object} httputils.HTTPError
// @success 200 {object} coreapi.DocumentResponse
// @router /documents/{document_id}/versions/{version_id} [get]
func (h handler) GetDocumentVersion(w http.ResponseWriter, r *http.Request) {
	var err error
	var code int
	defer httputils.RespondIfError(&code, &err, w, r)

	ids := make([][]byte, 2, 2)
	for i, idStr := range []string{chi.URLParam(r, "document_id"), chi.URLParam(r, "version_id")} {
		var id []byte
		id, err = hexutil.Decode(idStr)
		if err != nil {
			code = http.StatusBadRequest
			log.Error(err)
			err = ErrInvalidDocumentID
			return
		}

		ids[i] = id
	}

	model, err := h.srv.GetDocumentVersion(r.Context(), ids[0], ids[1])
	if err != nil {
		code = http.StatusNotFound
		log.Error(err)
		err = ErrDocumentNotFound
		return
	}

	resp, err := getDocumentResponse(model, h.tokenRegistry, jobs.NilJobID())
	if err != nil {
		code = http.StatusInternalServerError
		log.Error(err)
		return
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, resp)
}
