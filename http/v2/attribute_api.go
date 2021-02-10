package v2

import (
	"net/http"

	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/http/coreapi"
	"github.com/centrifuge/go-centrifuge/utils/httputils"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/go-chi/chi"
	"github.com/go-chi/render"
)

const (
	// AttributeKeyParam param in the URL.
	AttributeKeyParam = "attribute_key"

	// ErrInvalidAttributeKey for invalid attribute key in the api path.
	ErrInvalidAttributeKey = errors.Error("Invalid Attribute Key")
)

// SignedAttributeRequest contains the payload to be signed and added to the document.
type SignedAttributeRequest struct {
	Label   string `json:"label"`
	Type    string `json:"type" enums:"integer,string,bytes,timestamp"`
	Payload string `json:"payload"`
}

// AddSignedAttribute signs the given payload and add it the pending document.
// @summary Signs the given payload and add it the pending document.
// @description Signs the given payload and add it the pending document.
// @id add_signed_attribute
// @tags Documents
// @accept json
// @param authorization header string true "Hex encoded centrifuge ID of the account for the intended API action"
// @param body body v2.SignedAttributeRequest true "Signed Attribute request"
// @param document_id path string true "Document Identifier"
// @produce json
// @Failure 400 {object} httputils.HTTPError
// @Failure 500 {object} httputils.HTTPError
// @Failure 403 {object} httputils.HTTPError
// @success 200 {object} coreapi.DocumentResponse
// @router /v2/documents/{document_id}/signed_attribute [post]
func (h handler) AddSignedAttribute(w http.ResponseWriter, r *http.Request) {
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

	var req SignedAttributeRequest
	err = unmarshalBody(r, &req)
	if err != nil {
		code = http.StatusBadRequest
		log.Error(err)
		return
	}

	tp := documents.AttributeType(req.Type)
	val, err := documents.AttrValFromString(tp, req.Payload)
	if err != nil {
		code = http.StatusBadRequest
		log.Error(err)
		return
	}

	vb, err := val.ToBytes()
	if err != nil {
		code = http.StatusBadRequest
		log.Error(err)
		return
	}

	doc, err := h.srv.AddSignedAttribute(r.Context(), docID, req.Label, vb, tp)
	if err != nil {
		code = http.StatusBadRequest
		log.Error(err)
		return
	}

	resp, err := toDocumentResponse(doc, h.srv.tokenRegistry, "")
	if err != nil {
		code = http.StatusInternalServerError
		log.Error(err)
		return
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, resp)
}

// AddAttributes add attributes to pending document.
// @summary Add attributes to pending document.
// @description Add attributes to pending document.
// @id add_attributes
// @tags Documents
// @accept json
// @param authorization header string true "Hex encoded centrifuge ID of the account for the intended API action"
// @param body body coreapi.AttributeMapRequest true "Attributes"
// @param document_id path string true "Document Identifier"
// @produce json
// @Failure 400 {object} httputils.HTTPError
// @Failure 500 {object} httputils.HTTPError
// @Failure 403 {object} httputils.HTTPError
// @success 200 {object} coreapi.DocumentResponse
// @router /v2/documents/{document_id}/attributes [post]
func (h handler) AddAttributes(w http.ResponseWriter, r *http.Request) {
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

	var req coreapi.AttributeMapRequest
	err = unmarshalBody(r, &req)
	if err != nil {
		code = http.StatusBadRequest
		log.Error(err)
		return
	}

	attrs, err := toDocumentAttributes(req)
	if err != nil {
		code = http.StatusBadRequest
		log.Error(err)
		return
	}

	doc, err := h.srv.AddAttributes(r.Context(), docID, attrs)
	if err != nil {
		code = http.StatusBadRequest
		log.Error(err)
		return
	}

	resp, err := toDocumentResponse(doc, h.srv.tokenRegistry, "")
	if err != nil {
		code = http.StatusInternalServerError
		log.Error(err)
		return
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, resp)
}

// DeleteAttribute deletes attribute on a pending document.
// @summary Deletes attribute on a pending document.
// @description Deletes attribute on a pending document.
// @id delete_attribute
// @tags Documents
// @accept json
// @param authorization header string true "Hex encoded centrifuge ID of the account for the intended API action"
// @param document_id path string true "Document Identifier"
// @param attribute_key path string true "Hex encoded attribute key"
// @produce json
// @Failure 400 {object} httputils.HTTPError
// @Failure 500 {object} httputils.HTTPError
// @Failure 403 {object} httputils.HTTPError
// @success 200 {object} coreapi.DocumentResponse
// @router /v2/documents/{document_id}/attributes/{attribute_key} [delete]
func (h handler) DeleteAttribute(w http.ResponseWriter, r *http.Request) {
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

	attrKeyBytes, err := hexutil.Decode(chi.URLParam(r, AttributeKeyParam))
	if err != nil {
		code = http.StatusBadRequest
		log.Error(err)
		err = ErrInvalidAttributeKey
		return
	}

	attrKey, err := documents.AttrKeyFromBytes(attrKeyBytes)
	if err != nil {
		code = http.StatusBadRequest
		log.Error(err)
		err = ErrInvalidAttributeKey
		return
	}

	doc, err := h.srv.DeleteAttribute(r.Context(), docID, attrKey)
	if err != nil {
		code = http.StatusBadRequest
		log.Error(err)
		return
	}

	resp, err := toDocumentResponse(doc, h.srv.tokenRegistry, "")
	if err != nil {
		code = http.StatusInternalServerError
		log.Error(err)
		return
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, resp)
}
