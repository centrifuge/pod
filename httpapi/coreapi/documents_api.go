package coreapi

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/jobs"
	"github.com/centrifuge/go-centrifuge/utils/httputils"
	"github.com/ethereum/go-ethereum/common/hexutil"
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
// @param authorization header string true "Hex encoded centrifuge ID of the account for the intended API action"
// @param body body coreapi.CreateDocumentRequest true "Document Create request"
// @produce json
// @Failure 500 {object} httputils.HTTPError
// @Failure 400 {object} httputils.HTTPError
// @Failure 403 {object} httputils.HTTPError
// @success 202 {object} coreapi.DocumentResponse
// @router /v1/documents [post]
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

	payload, err := ToDocumentsCreatePayload(request)
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

	resp, err := GetDocumentResponse(model, h.tokenRegistry, jobID)
	if err != nil {
		code = http.StatusInternalServerError
		log.Error(err)
		return
	}

	render.Status(r, http.StatusAccepted)
	render.JSON(w, r, resp)
}

// UpdateDocument updates an existing document.
// @summary Updates an existing document and anchors it.
// @description Updates an existing document and anchors it.
// @id update_document
// @tags Documents
// @accept json
// @param authorization header string true "Hex encoded centrifuge ID of the account for the intended API action"
// @param document_id path string true "Document Identifier"
// @param body body coreapi.CreateDocumentRequest true "Document Update request"
// @produce json
// @Failure 500 {object} httputils.HTTPError
// @Failure 400 {object} httputils.HTTPError
// @Failure 403 {object} httputils.HTTPError
// @success 202 {object} coreapi.DocumentResponse
// @router /v1/documents/{document_id} [put]
func (h handler) UpdateDocument(w http.ResponseWriter, r *http.Request) {
	var err error
	var code int
	defer httputils.RespondIfError(&code, &err, w, r)

	docID, err := hexutil.Decode(chi.URLParam(r, DocumentIDParam))
	if err != nil {
		code = http.StatusBadRequest
		log.Error(err)
		err = ErrInvalidDocumentID
		return
	}

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

	payload, err := ToDocumentsCreatePayload(request)
	if err != nil {
		code = http.StatusBadRequest
		log.Error(err)
		return
	}

	model, jobID, err := h.srv.UpdateDocument(ctx, documents.UpdatePayload{DocumentID: docID, CreatePayload: payload})
	if err != nil {
		code = http.StatusBadRequest
		log.Error(err)
		return
	}

	resp, err := GetDocumentResponse(model, h.tokenRegistry, jobID)
	if err != nil {
		code = http.StatusInternalServerError
		log.Error(err)
		return
	}

	render.Status(r, http.StatusAccepted)
	render.JSON(w, r, resp)
}

// GetDocument returns the latest version of the document.
// @summary Returns the latest version of the document.
// @description Returns the latest version of the document.
// @id get_document
// @tags Documents
// @param authorization header string true "Hex encoded centrifuge ID of the account for the intended API action"
// @param document_id path string true "Document Identifier"
// @produce json
// @Failure 403 {object} httputils.HTTPError
// @Failure 400 {object} httputils.HTTPError
// @Failure 404 {object} httputils.HTTPError
// @Failure 500 {object} httputils.HTTPError
// @success 200 {object} coreapi.DocumentResponse
// @router /v1/documents/{document_id} [get]
func (h handler) GetDocument(w http.ResponseWriter, r *http.Request) {
	var err error
	var code int
	defer httputils.RespondIfError(&code, &err, w, r)

	docID, err := hexutil.Decode(chi.URLParam(r, DocumentIDParam))
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

	resp, err := GetDocumentResponse(model, h.tokenRegistry, jobs.NilJobID())
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
// @param authorization header string true "Hex encoded centrifuge ID of the account for the intended API action"
// @param document_id path string true "Document Identifier"
// @param version_id path string true "Document Version Identifier"
// @produce json
// @Failure 403 {object} httputils.HTTPError
// @Failure 400 {object} httputils.HTTPError
// @Failure 404 {object} httputils.HTTPError
// @Failure 500 {object} httputils.HTTPError
// @success 200 {object} coreapi.DocumentResponse
// @router /v1/documents/{document_id}/versions/{version_id} [get]
func (h handler) GetDocumentVersion(w http.ResponseWriter, r *http.Request) {
	var err error
	var code int
	defer httputils.RespondIfError(&code, &err, w, r)

	ids := make([][]byte, 2, 2)
	for i, idStr := range []string{chi.URLParam(r, DocumentIDParam), chi.URLParam(r, VersionIDParam)} {
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

	resp, err := GetDocumentResponse(model, h.tokenRegistry, jobs.NilJobID())
	if err != nil {
		code = http.StatusInternalServerError
		log.Error(err)
		return
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, resp)
}

// GenerateProofs returns proofs for the fields from latest version of the document.
// @summary Generates proofs for the fields from latest version of the document.
// @description Generates proofs for the fields from latest version of the document.
// @id generate_document_proofs
// @tags Documents
// @param authorization header string true "Hex encoded centrifuge ID of the account for the intended API action"
// @param document_id path string true "Document Identifier"
// @param body body coreapi.ProofsRequest true "Document proof request"
// @produce json
// @Failure 403 {object} httputils.HTTPError
// @Failure 400 {object} httputils.HTTPError
// @Failure 500 {object} httputils.HTTPError
// @success 200 {object} coreapi.ProofsResponse
// @router /v1/documents/{document_id}/proofs [post]
func (h handler) GenerateProofs(w http.ResponseWriter, r *http.Request) {
	var err error
	var code int
	defer httputils.RespondIfError(&code, &err, w, r)

	docID, err := hexutil.Decode(chi.URLParam(r, DocumentIDParam))
	if err != nil {
		code = http.StatusBadRequest
		log.Error(err)
		err = ErrInvalidDocumentID
		return
	}

	d, err := ioutil.ReadAll(r.Body)
	if err != nil {
		code = http.StatusInternalServerError
		log.Error(err)
		return
	}

	var request ProofsRequest
	err = json.Unmarshal(d, &request)
	if err != nil {
		code = http.StatusBadRequest
		log.Error(err)
		return
	}

	proofs, err := h.srv.GenerateProofs(r.Context(), docID, request.Fields)
	if err != nil {
		code = http.StatusInternalServerError
		log.Error(err)
		return
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, ConvertProofs(proofs))
}

// GenerateProofsForVersion returns proofs for the fields from a specific document version.
// @summary Generates proofs for the fields from a specific document version.
// @description Generates proofs for the fields from a specific document version.
// @id generate_document_version_proofs
// @tags Documents
// @param authorization header string true "Hex encoded centrifuge ID of the account for the intended API action"
// @param document_id path string true "Document Identifier"
// @param version_id path string true "Document Version Identifier"
// @param body body coreapi.ProofsRequest true "Document proof request"
// @produce json
// @Failure 403 {object} httputils.HTTPError
// @Failure 400 {object} httputils.HTTPError
// @Failure 500 {object} httputils.HTTPError
// @success 200 {object} coreapi.ProofsResponse
// @router /v1/documents/{document_id}/versions/{version_id}/proofs [post]
func (h handler) GenerateProofsForVersion(w http.ResponseWriter, r *http.Request) {
	var err error
	var code int
	defer httputils.RespondIfError(&code, &err, w, r)

	ids := make([][]byte, 2, 2)
	for i, idStr := range []string{chi.URLParam(r, DocumentIDParam), chi.URLParam(r, VersionIDParam)} {
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

	d, err := ioutil.ReadAll(r.Body)
	if err != nil {
		code = http.StatusInternalServerError
		log.Error(err)
		return
	}

	var request ProofsRequest
	err = json.Unmarshal(d, &request)
	if err != nil {
		code = http.StatusBadRequest
		log.Error(err)
		return
	}

	proofs, err := h.srv.GenerateProofsForVersion(r.Context(), ids[0], ids[1], request.Fields)
	if err != nil {
		code = http.StatusInternalServerError
		log.Error(err)
		return
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, ConvertProofs(proofs))
}
