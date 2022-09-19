package v2

import (
	"net/http"

	"github.com/centrifuge/go-substrate-rpc-client/v4/types"

	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/http/coreapi"
	"github.com/centrifuge/go-centrifuge/utils/byteutils"
	"github.com/centrifuge/go-centrifuge/utils/httputils"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/go-chi/chi"
	"github.com/go-chi/render"
)

// CreateDocumentRequest defines the payload for creating documents.
type CreateDocumentRequest struct {
	coreapi.CreateDocumentRequest
	DocumentID byteutils.OptionalHex `json:"document_id" swaggertype:"primitive,string"` // if provided, creates the next version of the document.
}

// CloneDocumentRequest defines the payload for creating documents.
type CloneDocumentRequest struct {
	Scheme string `json:"scheme" enums:"generic,entity"`
}

// UpdateDocumentRequest defines the payload to patch an existing document.
type UpdateDocumentRequest struct {
	coreapi.CreateDocumentRequest
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
	var req CreateDocumentRequest
	err = unmarshalBody(r, &req)
	if err != nil {
		code = http.StatusBadRequest
		log.Error(err)
		return
	}

	payload, err := toDocumentsPayload(req.CreateDocumentRequest, req.DocumentID.Bytes())
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

	resp, err := toDocumentResponse(doc, "")
	if err != nil {
		code = http.StatusInternalServerError
		log.Error(err)
		return
	}

	render.Status(r, http.StatusCreated)
	render.JSON(w, r, resp)
}

// CloneDocument creates a new cloned document from an existing Template document.
// @summary Creates a new cloned document from an existing Template document.
// @description Creates a new cloned document from an existing Template document.
// @id clone_document_v2
// @tags Documents
// @accept json
// @param authorization header string true "Hex encoded centrifuge ID of the account for the intended API action"
// @param body body v2.CloneDocumentRequest true "Document Clone request"
// @param document_id path string true "Document Identifier"
// @produce json
// @Failure 400 {object} httputils.HTTPError
// @Failure 500 {object} httputils.HTTPError
// @Failure 403 {object} httputils.HTTPError
// @success 201 {object} coreapi.DocumentResponse
// @router /v2/documents/{document_id}/clone [post]
func (h handler) CloneDocument(w http.ResponseWriter, r *http.Request) {
	var err error
	var code int
	defer httputils.RespondIfError(&code, &err, w, r)

	templateID, err := hexutil.Decode(chi.URLParam(r, coreapi.DocumentIDParam))
	if err != nil {
		code = http.StatusBadRequest
		log.Error(err)
		err = coreapi.ErrInvalidDocumentID
		return
	}

	ctx := r.Context()
	var req CloneDocumentRequest
	err = unmarshalBody(r, &req)
	if err != nil {
		code = http.StatusBadRequest
		log.Error(err)
		return
	}

	payload := documents.ClonePayload{
		Scheme:     req.Scheme,
		TemplateID: templateID,
	}

	doc, err := h.srv.CloneDocument(ctx, payload)
	if err != nil {
		code = http.StatusBadRequest
		log.Error(err)
		return
	}

	resp, err := toDocumentResponse(doc, "")
	if err != nil {
		code = http.StatusInternalServerError
		log.Error(err)
		return
	}

	render.Status(r, http.StatusCreated)
	render.JSON(w, r, resp)
}

// UpdateDocument updates a pending document.
// @summary Updates a pending document.
// @description Updates a pending document.
// @id update_document_v2
// @tags Documents
// @accept json
// @param authorization header string true "Hex encoded centrifuge ID of the account for the intended API action"
// @param body body v2.UpdateDocumentRequest true "Document Update request"
// @param document_id path string true "Document Identifier"
// @produce json
// @Failure 400 {object} httputils.HTTPError
// @Failure 500 {object} httputils.HTTPError
// @Failure 403 {object} httputils.HTTPError
// @success 200 {object} coreapi.DocumentResponse
// @router /v2/documents/{document_id} [patch]
func (h handler) UpdateDocument(w http.ResponseWriter, r *http.Request) {
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
	var req UpdateDocumentRequest
	err = unmarshalBody(r, &req)
	if err != nil {
		code = http.StatusBadRequest
		log.Error(err)
		return
	}

	payload, err := toDocumentsPayload(req.CreateDocumentRequest, docID)
	if err != nil {
		code = http.StatusBadRequest
		log.Error(err)
		return
	}

	doc, err := h.srv.UpdateDocument(ctx, payload)
	if err != nil {
		code = http.StatusBadRequest
		log.Error(err)
		return
	}

	resp, err := toDocumentResponse(doc, "")
	if err != nil {
		code = http.StatusInternalServerError
		log.Error(err)
		return
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, resp)
}

// Commit creates a document.
// @summary Commits a pending document.
// @description Commits a pending document.
// @id commit_document_v2
// @tags Documents
// @accept json
// @param authorization header string true "Hex encoded centrifuge ID of the account for the intended API action"
// @param document_id path string true "Document Identifier"
// @produce json
// @Failure 400 {object} httputils.HTTPError
// @Failure 500 {object} httputils.HTTPError
// @Failure 403 {object} httputils.HTTPError
// @success 202 {object} coreapi.DocumentResponse
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
	doc, jobID, err := h.srv.Commit(ctx, docID)
	if err != nil {
		code = http.StatusBadRequest
		log.Error(err)
		return
	}

	resp, err := toDocumentResponse(doc, jobID.Hex())
	if err != nil {
		code = http.StatusInternalServerError
		log.Error(err)
		return
	}
	render.Status(r, http.StatusAccepted)
	render.JSON(w, r, resp)
}

func (h handler) getDocumentWithStatus(w http.ResponseWriter, r *http.Request, st documents.Status) {
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
	doc, err := h.srv.GetDocument(ctx, docID, st)
	if err != nil {
		code = http.StatusNotFound
		log.Error(err)
		err = coreapi.ErrDocumentNotFound
		return
	}

	resp, err := toDocumentResponse(doc, "")
	if err != nil {
		code = http.StatusInternalServerError
		log.Error(err)
		return
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, resp)
}

// GetPendingDocument returns the pending document associated with docID.
// @summary Returns the pending document associated with docID.
// @description Returns the pending document associated with docID.
// @id get_pending_document
// @tags Documents
// @param authorization header string true "Hex encoded centrifuge ID of the account for the intended API action"
// @param document_id path string true "Document Identifier"
// @produce json
// @Failure 403 {object} httputils.HTTPError
// @Failure 400 {object} httputils.HTTPError
// @Failure 404 {object} httputils.HTTPError
// @Failure 500 {object} httputils.HTTPError
// @success 200 {object} coreapi.DocumentResponse
// @router /v2/documents/{document_id}/pending [get]
func (h handler) GetPendingDocument(w http.ResponseWriter, r *http.Request) {
	h.getDocumentWithStatus(w, r, documents.Pending)
}

// GetCommittedDocument returns the latest committed document associated with docID.
// @summary Returns the latest committed document associated with docID.
// @description Returns the latest committed document associated with docID.
// @id get_committed_document
// @tags Documents
// @param authorization header string true "Hex encoded centrifuge ID of the account for the intended API action"
// @param document_id path string true "Document Identifier"
// @produce json
// @Failure 403 {object} httputils.HTTPError
// @Failure 400 {object} httputils.HTTPError
// @Failure 404 {object} httputils.HTTPError
// @Failure 500 {object} httputils.HTTPError
// @success 200 {object} coreapi.DocumentResponse
// @router /v2/documents/{document_id}/committed [get]
func (h handler) GetCommittedDocument(w http.ResponseWriter, r *http.Request) {
	h.getDocumentWithStatus(w, r, documents.Committed)
}

// GetDocumentVersion returns the specific version of the document.
// @summary Returns the specific version of the document.
// @description Returns the specific version of the document.
// @id get_document_version_v2
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
// @router /v2/documents/{document_id}/versions/{version_id} [get]
func (h handler) GetDocumentVersion(w http.ResponseWriter, r *http.Request) {
	var err error
	var code int
	defer httputils.RespondIfError(&code, &err, w, r)

	ids := make([][]byte, 2)
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

	doc, err := h.srv.GetDocumentVersion(r.Context(), ids[0], ids[1])
	if err != nil {
		code = http.StatusNotFound
		log.Error(err)
		err = coreapi.ErrDocumentNotFound
		return
	}

	resp, err := toDocumentResponse(doc, "")
	if err != nil {
		code = http.StatusInternalServerError
		log.Error(err)
		return
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, resp)
}

// RemoveCollaboratorsRequest contains the list of collaborators that are to be removed from the document
type RemoveCollaboratorsRequest struct {
	Collaborators []*types.AccountID `json:"collaborators" swaggertype:"array,string"`
}

// RemoveCollaborators removes the collaborators from the document.
// @summary Removes the collaborators from the document.
// @description Removes the collaborators from the document.
// @id remove_collaborators
// @tags Documents
// @accept json
// @param authorization header string true "Hex encoded centrifuge ID of the account for the intended API action"
// @param body body v2.RemoveCollaboratorsRequest true "Remove Collaborators request"
// @param document_id path string true "Document Identifier"
// @produce json
// @Failure 400 {object} httputils.HTTPError
// @Failure 500 {object} httputils.HTTPError
// @Failure 403 {object} httputils.HTTPError
// @success 200 {object} coreapi.DocumentResponse
// @router /v2/documents/{document_id}/collaborators [delete]
func (h handler) RemoveCollaborators(w http.ResponseWriter, r *http.Request) {
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

	var req RemoveCollaboratorsRequest
	err = unmarshalBody(r, &req)
	if err != nil {
		code = http.StatusBadRequest
		log.Error(err)
		return
	}

	doc, err := h.srv.RemoveCollaborators(r.Context(), docID, req.Collaborators)
	if err != nil {
		code = http.StatusBadRequest
		log.Error(err)
		return
	}

	resp, err := toDocumentResponse(doc, "")
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
// @success 200 {object} coreapi.ProofsResponse
// @router /v2/documents/{document_id}/proofs [post]
func (h handler) GenerateProofs(w http.ResponseWriter, r *http.Request) {
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

	var req coreapi.ProofsRequest
	err = unmarshalBody(r, &req)
	if err != nil {
		code = http.StatusBadRequest
		log.Error(err)
		return
	}

	proofs, err := h.srv.GenerateProofs(r.Context(), docID, req.Fields)
	if err != nil {
		code = http.StatusBadRequest
		log.Error(err)
		return
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, coreapi.ConvertProofs(proofs))
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
// @router /v2/documents/{document_id}/versions/{version_id}/proofs [post]
func (h handler) GenerateProofsForVersion(w http.ResponseWriter, r *http.Request) {
	var err error
	var code int
	defer httputils.RespondIfError(&code, &err, w, r)

	ids := make([][]byte, 2)
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

	var req coreapi.ProofsRequest
	err = unmarshalBody(r, &req)
	if err != nil {
		code = http.StatusBadRequest
		log.Error(err)
		return
	}

	proofs, err := h.srv.GenerateProofsForVersion(r.Context(), ids[0], ids[1], req.Fields)
	if err != nil {
		code = http.StatusBadRequest
		log.Error(err)
		return
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, coreapi.ConvertProofs(proofs))
}
