package v2

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/centrifuge/go-centrifuge/http/coreapi"
	"github.com/centrifuge/go-centrifuge/nft"
	"github.com/centrifuge/go-centrifuge/utils/httputils"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/go-chi/chi"
	"github.com/go-chi/render"
)

// MintNFTOnCC mints an NFT on centrifuge chain.
// @summary Mints an NFT against a document on centrifuge chain.
// @description Mints an NFT against a document on centrifuge chain.
// @id mint_nft_cc
// @tags NFTs
// @accept json
// @param authorization header string true "Hex encoded centrifuge ID of the account for the intended API action"
// @param registry_address path string true "NFT registry address in hex"
// @param body body coreapi.MintNFTOnCCRequest true "Mint NFT on CC request"
// @produce json
// @Failure 403 {object} httputils.HTTPError
// @Failure 500 {object} httputils.HTTPError
// @Failure 400 {object} httputils.HTTPError
// @success 202 {object} coreapi.MintNFTOnCCResponse
// @router /beta/nfts/registries/{registry_address}/mint [post]
func (h handler) MintNFTOnCC(w http.ResponseWriter, r *http.Request) {
	var err error
	var code int
	defer httputils.RespondIfError(&code, &err, w, r)

	if !common.IsHexAddress(chi.URLParam(r, coreapi.RegistryAddressParam)) {
		code = http.StatusBadRequest
		err = ErrInvalidRegistryAddress
		log.Error(err)
		return
	}

	registry := common.HexToAddress(chi.URLParam(r, coreapi.RegistryAddressParam))
	ctx := r.Context()
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		code = http.StatusInternalServerError
		log.Error(err)
		return
	}

	var req coreapi.MintNFTOnCCRequest
	err = json.Unmarshal(data, &req)
	if err != nil {
		code = http.StatusBadRequest
		log.Error(err)
		return
	}

	resp, err := h.srv.MintNFTOnCC(ctx, coreapi.ToNFTMintRequestOnCC(req, registry))
	if err != nil {
		code = http.StatusBadRequest
		log.Error(err)
		return
	}

	nftResp := coreapi.MintNFTOnCCResponse{
		Header:          coreapi.NFTResponseHeader{JobID: resp.JobID},
		RegistryAddress: registry,
		DepositAddress:  req.DepositAddress,
		DocumentID:      req.DocumentID,
		TokenID:         resp.TokenID,
	}
	render.Status(r, http.StatusAccepted)
	render.JSON(w, r, nftResp)
}

// TransferNFTOnCC transfers given NFT to provide address on centrifuge chain.
// @summary Transfers given NFT to provide address on centrifuge chain.
// @description Transfers given NFT to provide address on centrifuge chain.
// @id transfer_nft_cc
// @tags NFTs
// @accept json
// @param authorization header string true "Hex encoded centrifuge ID of the account for the intended API action"
// @param registry_address path string true "NFT registry address in hex"
// @param token_id path string true "NFT token ID in hex"
// @param body body coreapi.TransferNFTOnCCRequest true "Transfer NFT request"
// @produce json
// @Failure 403 {object} httputils.HTTPError
// @Failure 500 {object} httputils.HTTPError
// @Failure 400 {object} httputils.HTTPError
// @success 202 {object} coreapi.TransferNFTOnCCResponse
// @router /beta/nfts/registries/{registry_address}/tokens/{token_id}/transfer [post]
func (h handler) TransferNFTOnCC(w http.ResponseWriter, r *http.Request) {
	var err error
	var code int
	defer httputils.RespondIfError(&code, &err, w, r)

	if !common.IsHexAddress(chi.URLParam(r, coreapi.RegistryAddressParam)) {
		code = http.StatusBadRequest
		err = ErrInvalidRegistryAddress
		log.Error(err)
		return
	}

	registry := common.HexToAddress(chi.URLParam(r, coreapi.RegistryAddressParam))
	tokenID, err := nft.TokenIDFromString(chi.URLParam(r, coreapi.TokenIDParam))
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

	var req coreapi.TransferNFTOnCCRequest
	err = json.Unmarshal(data, &req)
	if err != nil {
		code = http.StatusBadRequest
		log.Error(err)
		return
	}

	toAccID, err := types.NewAccountID(req.To)

	if err != nil {
		code = http.StatusBadRequest
		log.Error(err)
		return
	}

	resp, err := h.srv.TransferNFTOnCC(ctx, registry, tokenID, *toAccID)
	if err != nil {
		code = http.StatusBadRequest
		log.Error(err)
		return
	}

	render.Status(r, http.StatusAccepted)
	render.JSON(w, r, coreapi.TransferNFTOnCCResponse{
		RegistryAddress: registry,
		To:              req.To,
		TokenID:         resp.TokenID,
		Header:          coreapi.NFTResponseHeader{JobID: resp.JobID},
	})
}

// OwnerOfNFTOnCC returns the owner of the given NFT on centrifuge chain.
// @summary Returns the Owner of the given NFT on centrifuge chain.
// @description Returns the Owner of the given NFT on centrifuge chain.
// @id owner_of_nft_cc
// @tags NFTs
// @param authorization header string true "Hex encoded centrifuge ID of the account for the intended API action"
// @param token_id path string true "NFT token ID in hex"
// @param registry_address path string true "Registry address in hex"
// @produce json
// @Failure 403 {object} httputils.HTTPError
// @Failure 400 {object} httputils.HTTPError
// @success 200 {object} coreapi.NFTOwnerOnCCResponse
// @router /beta/nfts/registries/{registry_address}/tokens/{token_id}/owner [get]
func (h handler) OwnerOfNFTOnCC(w http.ResponseWriter, r *http.Request) {
	var err error
	var code int
	defer httputils.RespondIfError(&code, &err, w, r)

	if !common.IsHexAddress(chi.URLParam(r, coreapi.RegistryAddressParam)) {
		code = http.StatusBadRequest
		err = ErrInvalidRegistryAddress
		log.Error(err)
		return
	}

	tokenID, err := nft.TokenIDFromString(chi.URLParam(r, coreapi.TokenIDParam))
	if err != nil {
		code = http.StatusBadRequest
		log.Error(err)
		err = ErrInvalidTokenID
		return
	}

	registry := common.HexToAddress(chi.URLParam(r, coreapi.RegistryAddressParam))
	owner, err := h.srv.OwnerOfNFTOnCC(registry, tokenID)
	if err != nil {
		code = http.StatusBadRequest
		log.Error(err)
		return
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, coreapi.NFTOwnerOnCCResponse{
		TokenID:         tokenID.String(),
		RegistryAddress: registry,
		Owner:           owner[:],
	})
}
