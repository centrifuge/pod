package coreapi

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/utils/httputils"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/go-chi/chi"
	"github.com/go-chi/render"
)

const (
	// ErrAccountIDInvalid is a sentinel error for invalid account IDs.
	ErrAccountIDInvalid = errors.Error("account ID is invalid")

	// ErrAccountNotFound is a sentinel error for when account is missing.
	ErrAccountNotFound = errors.Error("account not found")
)

// SignPayload signs the payload and returns the signature.
// @summary Signs and returns the signature of the Payload.
// @description Signs and returns the signature of the Payload.
// @id account_sign
// @tags Accounts
// @param account_id path string true "Account ID"
// @param body body coreapi.SignRequest true "Sign request"
// @produce json
// @Failure 400 {object} httputils.HTTPError
// @Failure 500 {object} httputils.HTTPError
// @success 200 {object} coreapi.SignResponse
// @router /v1/accounts/{account_id}/sign [post]
func (h handler) SignPayload(w http.ResponseWriter, r *http.Request) {
	var err error
	var code int
	defer httputils.RespondIfError(&code, &err, w, r)

	accID, err := hexutil.Decode(chi.URLParam(r, accountIDParam))
	if err != nil {
		code = http.StatusBadRequest
		log.Error(err)
		err = ErrAccountIDInvalid
		return
	}

	d, err := ioutil.ReadAll(r.Body)
	if err != nil {
		code = http.StatusInternalServerError
		log.Error(err)
		return
	}

	var payload SignRequest
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
	render.JSON(w, r, SignResponse{
		Payload:   payload.Payload,
		PublicKey: sig.PublicKey,
		Signature: sig.Signature,
		SignerID:  sig.SignerId,
	})
}

// GetAccount returns the account associated with accountID.
// @summary Returns the account associated with accountID.
// @description Returns the account associated with accountID.
// @id get_account
// @tags Accounts
// @param account_id path string true "Account ID"
// @produce json
// @Failure 400 {object} httputils.HTTPError
// @Failure 404 {object} httputils.HTTPError
// @success 200 {object} coreapi.Account
// @router /v1/accounts/{account_id} [get]
func (h handler) GetAccount(w http.ResponseWriter, r *http.Request) {
	var err error
	var code int
	defer httputils.RespondIfError(&code, &err, w, r)

	accID, err := hexutil.Decode(chi.URLParam(r, accountIDParam))
	if err != nil {
		code = http.StatusBadRequest
		log.Error(err)
		err = ErrAccountIDInvalid
		return
	}

	acc, err := h.srv.GetAccount(accID)
	if err != nil {
		code = http.StatusNotFound
		log.Error(err)
		err = ErrAccountNotFound
		return
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, toClientAccount(acc))
}

// GenerateAccount generates a new account with defaults.
// @summary Generates a new account with defaults.
// @description Generates a new account with defaults.
// @id generate_account
// @tags Accounts
// @produce json
// @param body body config.CentChainAccount true "Centrifuge Chain Account"
// @Failure 400 {object} httputils.HTTPError
// @Failure 500 {object} httputils.HTTPError
// @success 200 {object} coreapi.Account
// @router /v1/accounts/generate [post]
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

	var cacc config.CentChainAccount
	err = json.Unmarshal(data, &cacc)
	if err != nil {
		code = http.StatusBadRequest
		log.Error(err)
		return
	}

	acc, err := h.srv.GenerateAccount(cacc)
	if err != nil {
		code = http.StatusInternalServerError
		log.Error(err)
		return
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, toClientAccount(acc))
}

// GetAccounts returns all the accounts in the node.
// @summary Returns all the accounts in the node.
// @description Returns all the accounts in the node.
// @id get_accounts
// @tags Accounts
// @produce json
// @Failure 500 {object} httputils.HTTPError
// @success 200 {object} coreapi.Accounts
// @router /v1/accounts [get]
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

	render.Status(r, http.StatusOK)
	render.JSON(w, r, toClientAccounts(accs))
}

// CreateAccount creates a new account.
// @summary Creates a new account without any default configurations.
// @description Creates a new account without any default configurations.
// @id create_account
// @tags Accounts
// @produce json
// @param body body coreapi.Account true "Account Create request"
// @Failure 400 {object} httputils.HTTPError
// @Failure 500 {object} httputils.HTTPError
// @success 200 {object} coreapi.Account
// @router /v1/accounts [post]
func (h handler) CreateAccount(w http.ResponseWriter, r *http.Request) {
	var err error
	var code int
	defer httputils.RespondIfError(&code, &err, w, r)

	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		code = http.StatusInternalServerError
		log.Error(err)
		return
	}

	var cacc Account
	err = json.Unmarshal(data, &cacc)
	if err != nil {
		code = http.StatusBadRequest
		log.Error(err)
		return
	}

	acc, err := fromClientAccount(cacc)
	if err != nil {
		code = http.StatusBadRequest
		log.Error(err)
		return
	}

	acc, err = h.srv.CreateAccount(acc)
	if err != nil {
		code = http.StatusInternalServerError
		log.Error(err)
		return
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, toClientAccount(acc))
}

// UpdateAccount updates an existing account.
// @summary Updates an existing account.
// @description Updates an existing account.
// @id update_account
// @tags Accounts
// @produce json
// @param account_id path string true "Account ID"
// @param body body coreapi.Account true "Account Update request"
// @Failure 400 {object} httputils.HTTPError
// @Failure 404 {object} httputils.HTTPError
// @Failure 500 {object} httputils.HTTPError
// @success 200 {object} coreapi.Account
// @router /v1/accounts/{account_id} [put]
func (h handler) UpdateAccount(w http.ResponseWriter, r *http.Request) {
	var err error
	var code int
	defer httputils.RespondIfError(&code, &err, w, r)

	accID, err := hexutil.Decode(chi.URLParam(r, accountIDParam))
	if err != nil {
		code = http.StatusBadRequest
		log.Error(err)
		err = ErrAccountIDInvalid
		return
	}

	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		code = http.StatusInternalServerError
		log.Error(err)
		return
	}

	var cacc Account
	err = json.Unmarshal(data, &cacc)
	if err != nil {
		code = http.StatusBadRequest
		log.Error(err)
		return
	}

	cacc.IdentityID = accID
	acc, err := fromClientAccount(cacc)
	if err != nil {
		code = http.StatusBadRequest
		log.Error(err)
		return
	}

	acc, err = h.srv.UpdateAccount(acc)
	if err != nil {
		code = http.StatusNotFound
		log.Error(err)
		err = ErrAccountNotFound
		return
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, toClientAccount(acc))
}
