package v2

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/centrifuge/go-centrifuge/httpapi/coreapi"
	"github.com/centrifuge/go-centrifuge/utils/byteutils"
	"github.com/centrifuge/go-centrifuge/utils/httputils"
	"github.com/go-chi/render"
)

// GenerateAccountResponse contains the expected DID and the jobID associated with the create identity Job
type GenerateAccountResponse struct {
	DID   byteutils.HexBytes `json:"did" swaggertype:"primitive,string"`
	JobID byteutils.HexBytes `json:"job_id" swaggertype:"primitive,string"`
}

// GenerateAccount generates a new account with defaults.
// @summary Generates a new account with defaults.
// @description Generates a new account with defaults.
// @id generate_account_v2
// @tags Accounts
// @produce json
// @param body body coreapi.GenerateAccountPayload true "Generate Account Payload"
// @Failure 400 {object} httputils.HTTPError
// @Failure 500 {object} httputils.HTTPError
// @success 201 {object} v2.GenerateAccountResponse
// @router /v2/accounts/generate [post]
func (h handler) GenerateAccount(w http.ResponseWriter, r *http.Request) {
	var err error
	var code int
	defer httputils.RespondIfError(&code, &err, w, r)

	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		code = http.StatusInternalServerError
		log.Error(err)
		return
	}

	var payload coreapi.GenerateAccountPayload
	err = json.Unmarshal(data, &payload)
	if err != nil {
		code = http.StatusBadRequest
		log.Error(err)
		return
	}

	did, jobID, err := h.srv.GenerateAccount(payload.CentChainAccount)
	if err != nil {
		code = http.StatusInternalServerError
		log.Error(err)
		return
	}

	render.Status(r, http.StatusCreated)
	render.JSON(w, r, GenerateAccountResponse{
		DID:   did,
		JobID: jobID,
	})
}
