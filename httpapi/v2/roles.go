package v2

import (
	"net/http"

	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/httpapi/coreapi"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/utils/byteutils"
	"github.com/centrifuge/go-centrifuge/utils/httputils"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/go-chi/chi"
	"github.com/go-chi/render"
)

// RoleIDParam is the key for roleID in the API path.
const RoleIDParam = "role_id"

// ErrInvalidRoleID for invalid roleID in the api path.
const ErrInvalidRoleID = errors.Error("Invalid RoleID")

// Role is a single role in the document.
type Role struct {
	ID            byteutils.HexBytes   `json:"id" swaggertype:"primitive,string"`
	Collaborators []byteutils.HexBytes `json:"collaborators" swaggertype:"array,string"`
}

// AddRole used for marshalling add request for role.
type AddRole struct {
	// Key is either hex encoded 32 byte ID or string label.
	// String label is used as a preimage to sha256 for 32 byte hash.
	Key           string         `json:"key"`
	Collaborators []identity.DID `json:"collaborators" swaggertype:"array,string"`
}

// GetRole returns the role associated with the role ID in the document
// @summary Returns the role associated with the role ID in the document.
// @description Returns the role associated with the role ID in the document.
// @id get_role
// @tags Documents
// @param authorization header string true "Hex encoded centrifuge ID of the account for the intended API action"
// @param document_id path string true "Document Identifier"
// @param role_id path string true "Role ID"
// @produce json
// @Failure 403 {object} httputils.HTTPError
// @Failure 400 {object} httputils.HTTPError
// @Failure 404 {object} httputils.HTTPError
// @success 200 {object} v2.Role
// @router /v2/documents/{document_id}/roles/{role_id} [get]
func (h handler) GetRole(w http.ResponseWriter, r *http.Request) {
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

	roleID, err := hexutil.Decode(chi.URLParam(r, RoleIDParam))
	if err != nil {
		code = http.StatusBadRequest
		log.Error(err)
		err = ErrInvalidRoleID
		return
	}

	rl, err := h.srv.GetRole(r.Context(), docID, roleID)
	if err != nil {
		code = http.StatusNotFound
		log.Error(err)
		return
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, toClientRole(rl))
}

// AddRle adds a new role to the document.
// @summary Adds a new role to the document.
// @description Adds a new role to the document.
// @id add_role
// @tags Documents
// @param authorization header string true "Hex encoded centrifuge ID of the account for the intended API action"
// @param document_id path string true "Document Identifier"
// @produce json
// @Failure 403 {object} httputils.HTTPError
// @Failure 400 {object} httputils.HTTPError
// @Failure 404 {object} httputils.HTTPError
// @success 200 {object} v2.Role
// @router /v2/documents/{document_id}/roles [post]
func (h handler) AddRole(w http.ResponseWriter, r *http.Request) {
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
	var rl AddRole
	err = unmarshalBody(r, &rl)
	if err != nil {
		code = http.StatusBadRequest
		log.Error(err)
		return
	}

	nrl, err := h.srv.AddRole(ctx, docID, rl.Key, rl.Collaborators)
	if err != nil {
		code = http.StatusBadRequest
		log.Error(err)
		return
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, toClientRole(nrl))
}
