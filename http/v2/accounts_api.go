package v2

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/centrifuge/go-centrifuge/contextutil"
	"github.com/centrifuge/go-centrifuge/http/coreapi"
	"github.com/centrifuge/go-centrifuge/utils/httputils"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/go-chi/chi"
	"github.com/go-chi/render"
)

// GenerateAccount generates a new account with defaults.
// @summary Generates a new account with defaults.
// @description Generates a new account with defaults.
// @id generate_account_v2
// @tags Accounts
// @produce json
// @param body body coreapi.GenerateAccountPayload true "Generate Account Payload"
// @Failure 400 {object} httputils.HTTPError
// @Failure 500 {object} httputils.HTTPError
// @success 201 {object} coreapi.Account
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

	account, err := h.srv.GenerateAccount(r.Context(), payload.ToCreateIdentityRequest())
	if err != nil {
		code = http.StatusInternalServerError
		log.Error(err)
		return
	}

	res, err := h.srv.ToClientAccounts(account)
	if err != nil {
		code = http.StatusInternalServerError
		log.Error(err)
		return
	}

	render.Status(r, http.StatusCreated)
	render.JSON(w, r, res[0])
}

// SignPayload signs the payload and returns the signature.
// @summary Signs and returns the signature of the Payload.
// @description Signs and returns the signature of the Payload.
// @id account_sign_v2
// @tags Accounts
// @param account_id path string true "Account ID"
// @param body body coreapi.SignRequest true "Sign request"
// @produce json
// @Failure 400 {object} httputils.HTTPError
// @Failure 500 {object} httputils.HTTPError
// @success 200 {object} coreapi.SignResponse
// @router /v2/accounts/{account_id}/sign [post]
func (h handler) SignPayload(w http.ResponseWriter, r *http.Request) {
	var err error
	var code int
	defer httputils.RespondIfError(&code, &err, w, r)

	accID, err := hexutil.Decode(chi.URLParam(r, coreapi.AccountIDParam))
	if err != nil {
		code = http.StatusBadRequest
		log.Error(err)
		err = coreapi.ErrAccountIDInvalid
		return
	}

	d, err := ioutil.ReadAll(r.Body)
	if err != nil {
		code = http.StatusInternalServerError
		log.Error(err)
		return
	}

	var payload coreapi.SignRequest
	err = json.Unmarshal(d, &payload)
	if err != nil {
		code = http.StatusBadRequest
		log.Error(err)
		return
	}

	sig, err := h.srv.SignPayload(accID, payload.Payload)
	if err != nil {
		code = http.StatusBadRequest
		log.Error(err)
		return
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, coreapi.SignResponse{
		Payload:   payload.Payload,
		PublicKey: sig.PublicKey,
		Signature: sig.Signature,
		SignerID:  sig.SignerId,
	})
}

// GetSelf returns the account associated with the identity provided in the JW3T auth token.
// @summary Returns the account associated with the identity provided in the JW3T auth token.
// @description Returns the account associated with the identity provided in the JW3T auth token.
// @id get_self_v2
// @tags Accounts
// @produce json
// @Failure 400 {object} httputils.HTTPError
// @Failure 404 {object} httputils.HTTPError
// @success 200 {object} coreapi.Account
// @router /v2/accounts/self [get]
func (h handler) GetSelf(w http.ResponseWriter, r *http.Request) {
	var err error
	var code int
	defer httputils.RespondIfError(&code, &err, w, r)

	acc, err := contextutil.Account(r.Context())
	if err != nil {
		code = http.StatusNotFound
		log.Error(err)
		err = coreapi.ErrAccountNotFound
		return
	}

	res, err := h.srv.ToClientAccounts(acc)
	if err != nil {
		code = http.StatusInternalServerError
		log.Error(err)
		return
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, res[0])
}

// GetAccount returns the account associated with accountID.
// @summary Returns the account associated with accountID.
// @description Returns the account associated with accountID.
// @id get_account_v2
// @tags Accounts
// @param account_id path string true "Account ID"
// @produce json
// @Failure 400 {object} httputils.HTTPError
// @Failure 404 {object} httputils.HTTPError
// @success 200 {object} coreapi.Account
// @router /v2/accounts/{account_id} [get]
func (h handler) GetAccount(w http.ResponseWriter, r *http.Request) {
	var err error
	var code int
	defer httputils.RespondIfError(&code, &err, w, r)

	accID, err := hexutil.Decode(chi.URLParam(r, coreapi.AccountIDParam))
	if err != nil {
		code = http.StatusBadRequest
		log.Error(err)
		err = coreapi.ErrAccountIDInvalid
		return
	}

	acc, err := h.srv.GetAccount(accID)
	if err != nil {
		code = http.StatusNotFound
		log.Error(err)
		err = coreapi.ErrAccountNotFound
		return
	}

	res, err := h.srv.ToClientAccounts(acc)
	if err != nil {
		code = http.StatusInternalServerError
		log.Error(err)
		return
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, res[0])
}

// GetAccounts returns all the accounts in the node.
// @summary Returns all the accounts in the node.
// @description Returns all the accounts in the node.
// @id get_accounts_v2
// @tags Accounts
// @produce json
// @Failure 500 {object} httputils.HTTPError
// @success 200 {object} coreapi.Accounts
// @router /v2/accounts [get]
func (h handler) GetAccounts(w http.ResponseWriter, r *http.Request) {
	var err error
	var code int
	defer httputils.RespondIfError(&code, &err, w, r)

	accs, err := h.srv.GetAccounts()
	if err != nil {
		code = http.StatusInternalServerError
		log.Error(err)
		return
	}

	res, err := h.srv.ToClientAccounts(accs...)
	if err != nil {
		code = http.StatusInternalServerError
		log.Error(err)
		return
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, coreapi.Accounts{Data: res})
}
