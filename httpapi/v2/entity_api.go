package v2

import (
	"net/http"

	"github.com/centrifuge/go-centrifuge/httpapi/coreapi"
	"github.com/centrifuge/go-centrifuge/jobs"
	"github.com/centrifuge/go-centrifuge/utils/httputils"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/go-chi/chi"
	"github.com/go-chi/render"
)

// GetEntityThroughRelationship returns the latest version of the Entity through relationship ID.
// @summary Returns the latest version of the Entity through relationship ID.
// @description Returns the latest version of the Entity through relationship ID.
// @id get_entity_through_relationship_id
// @tags Documents
// @param authorization header string true "Hex encoded centrifuge ID of the account for the intended API action"
// @param document_id path string true "Entity Relationship Document Identifier"
// @produce json
// @Failure 403 {object} httputils.HTTPError
// @Failure 400 {object} httputils.HTTPError
// @Failure 404 {object} httputils.HTTPError
// @Failure 500 {object} httputils.HTTPError
// @success 200 {object} coreapi.DocumentResponse
// @router /v2/relationships/{document_id}/entity [get]
func (h handler) GetEntityThroughRelationship(w http.ResponseWriter, r *http.Request) {
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
	entity, err := h.srv.GetEntityByRelationship(ctx, docID)
	if err != nil {
		code = http.StatusNotFound
		log.Error(err)
		err = coreapi.ErrDocumentNotFound
		return
	}

	resp, err := toDocumentResponse(entity, h.srv.tokenRegistry, jobs.NilJobID())
	if err != nil {
		code = http.StatusInternalServerError
		log.Error(err)
		return
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, resp)
}

// GetEntityRelationships returns the entity relationships.
// @summary Returns the entity relationships.
// @description Returns the entity relationships.
// @id get_entity_relationships
// @tags Documents
// @param authorization header string true "Hex encoded centrifuge ID of the account for the intended API action"
// @param document_id path string true "Document Identifier"
// @produce json
// @Failure 403 {object} httputils.HTTPError
// @Failure 400 {object} httputils.HTTPError
// @Failure 404 {object} httputils.HTTPError
// @Failure 500 {object} httputils.HTTPError
// @success 200 {object} []coreapi.DocumentResponse
// @router /v2/entities/{document_id}/relationships [get]
func (h handler) GetEntityRelationships(w http.ResponseWriter, r *http.Request) {
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
	relationships, err := h.srv.GetEntityRelationShips(ctx, docID)
	if err != nil {
		code = http.StatusNotFound
		log.Error(err)
		err = coreapi.ErrDocumentNotFound
		return
	}

	responses := make([]coreapi.DocumentResponse, len(relationships))
	for i, relationship := range relationships {
		resp, err := toDocumentResponse(relationship, h.srv.tokenRegistry, jobs.NilJobID())
		if err != nil {
			code = http.StatusInternalServerError
			log.Error(err)
			return
		}

		responses[i] = resp
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, responses)
}
