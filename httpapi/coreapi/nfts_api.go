package coreapi

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/nft"
	"github.com/centrifuge/go-centrifuge/utils/httputils"
	"github.com/go-chi/chi"
	"github.com/go-chi/render"
)

// ErrInvalidTokenID is a sentinel error when token ID is invalid
const ErrInvalidTokenID = errors.Error("Invalid Token ID")

// MintNFT mints an NFT.
// @summary Mints an NFT against a document.
// @description Mints an NFT against a document.
// @id mint_nft
// @tags NFT
// @accept json
// @param authorization header string true "centrifuge identity"
// @param body body coreapi.MintNFTRequest true "Mint NFT request"
// @produce json
// @Failure 500 {object} httputils.HTTPError
// @Failure 400 {object} httputils.HTTPError
// @success 201 {object} coreapi.MintNFTResponse
// @router /nfts/mint [post]
func (h handler) MintNFT(w http.ResponseWriter, r *http.Request) {
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

	var req MintNFTRequest
	err = json.Unmarshal(data, &req)
	if err != nil {
		code = http.StatusBadRequest
		log.Error(err)
		return
	}

	resp, err := h.srv.MintNFT(ctx, toNFTMintRequest(req))
	if err != nil {
		code = http.StatusBadRequest
		log.Error(err)
		return
	}

	nftResp := MintNFTResponse{
		Header:          NFTResponseHeader{JobID: resp.JobID},
		RegistryAddress: req.RegistryAddress,
		DepositAddress:  req.DepositAddress,
		DocumentID:      req.DocumentID,
		TokenID:         resp.TokenID,
	}
	render.Status(r, http.StatusCreated)
	render.JSON(w, r, nftResp)
}

// TransferNFT transfers given NFT to provide address.
// @summary Transfers given NFT to provide address.
// @description Transfers given NFT to provide address.
// @id transfer_nft
// @tags NFT
// @accept json
// @param authorization header string true "centrifuge identity"
// @param token_id path string true "NFT token ID in hex"
// @param body body coreapi.TransferNFTRequest true "Mint NFT request"
// @produce json
// @Failure 500 {object} httputils.HTTPError
// @Failure 400 {object} httputils.HTTPError
// @success 200 {object} coreapi.TransferNFTResponse
// @router /nfts/{token_id}/transfer [post]
func (h handler) TransferNFT(w http.ResponseWriter, r *http.Request) {
	var err error
	var code int
	defer httputils.RespondIfError(&code, &err, w, r)

	tokenID, err := nft.TokenIDFromString(chi.URLParam(r, tokenIDParam))
	if err != nil {
		code = http.StatusBadRequest
		log.Error(err)
		err = ErrInvalidTokenID
		return
	}

	ctx := r.Context()
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		code = http.StatusInternalServerError
		log.Error(err)
		return
	}

	var req TransferNFTRequest
	err = json.Unmarshal(data, &req)
	if err != nil {
		code = http.StatusBadRequest
		log.Error(err)
		return
	}

	resp, err := h.srv.TransferNFT(ctx, req.To, req.RegistryAddress, tokenID)
	if err != nil {
		code = http.StatusBadRequest
		log.Error(err)
		return
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, TransferNFTResponse{
		RegistryAddress: req.RegistryAddress,
		To:              req.To,
		TokenID:         resp.TokenID,
		Header:          NFTResponseHeader{JobID: resp.JobID},
	})
}
