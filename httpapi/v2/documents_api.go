package v2

import (
	"net/http"

	"github.com/centrifuge/go-centrifuge/httpapi/coreapi"
	"github.com/centrifuge/go-centrifuge/jobs"
	"github.com/centrifuge/go-centrifuge/utils/byteutils"
	"github.com/centrifuge/go-centrifuge/utils/httputils"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/go-chi/chi"
	"github.com/go-chi/render"
)

// DocumentRequest is an alias to coreapi Document request.
// Aliased here to fix the swagger generation issues.
type DocumentRequest = coreapi.CreateDocumentRequest

// CreateDocumentRequest defines the payload for creating documents.
type CreateDocumentRequest struct {
	DocumentRequest
	DocumentID byteutils.OptionalHex `json:"document_id" swaggertype:"primitive,string"` // if provided, creates the next version of the document.
}

// CreateDocument creates a document.
// @summary Creates a new document.
// @description Creates a new document.
// @id create_document_v2
// @tags Documents
// @accept json
// @param authorization header string true "Hex encoded centrifuge ID of the account for the intended API action"
// @param body body v2.CreateDocumentRequest true "Document Create request"
// @produce json
// @Failure 400 {object} httputils.HTTPError
// @Failure 500 {object} httputils.HTTPError
// @Failure 403 {object} httputils.HTTPError
// @success 201 {object} coreapi.DocumentResponse
// @router /v2/documents [post]
func (h handler) CreateDocument(w http.ResponseWriter, r *http.Request) {
	var err error
	var code int
	defer httputils.RespondIfError(&code, &err, w, r)

	ctx := r.Context()
	var request CreateDocumentRequest
	err = unmarshalBody(r, &request)
	if err != nil {
		code = http.StatusBadRequest
		log.Error(err)
		return
	}

	payload, err := toDocumentsPayload(request)
	if err != nil {
		code = http.StatusBadRequest
		log.Error(err)
		return
	}

	doc, err := h.srv.CreateDocument(ctx, payload)
	if err != nil {
		code = http.StatusBadRequest
		log.Error(err)
		return
	}

	resp, err := toDocumentResponse(doc, h.srv.tokenRegistry, jobs.NilJobID())
	if err != nil {
		code = http.StatusInternalServerError
		log.Error(err)
		return
	}

	render.Status(r, http.StatusCreated)
	render.JSON(w, r, resp)
}

// Commit creates a document.
// @summary Commits a pending document.
// @description Commits a pending document.
// @id commit_document_v2
// @tags Documents
// @accept json
// @param authorization header string true "Hex encoded centrifuge ID of the account for the intended API action"
// @param body body v2.CreateDocumentRequest true "Document Create request"
// @produce json
// @Failure 400 {object} httputils.HTTPError
// @Failure 500 {object} httputils.HTTPError
// @Failure 403 {object} httputils.HTTPError
// @success 201 {object} coreapi.DocumentResponse
// @router /v2/documents/{document_id}/commit [post]
func (h handler) Commit(w http.ResponseWriter, r *http.Request) {
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
	model, jobID, err := h.srv.Commit(ctx, docID)
	if err != nil {
		code = http.StatusBadRequest
		log.Error(err)
		return
	}

	resp, err := toDocumentResponse(model, h.srv.tokenRegistry, jobID)
	if err != nil {
		code = http.StatusInternalServerError
		log.Error(err)
		return
	}

	render.Status(r, http.StatusCreated)
	render.JSON(w, r, resp)
}
