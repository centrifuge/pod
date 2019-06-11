package coreapi

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/centrifuge/go-centrifuge/utils/httputils"
	"github.com/go-chi/render"
)

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
// @success 200 {object} coreapi.MintNFTResponse
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
