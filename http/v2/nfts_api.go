package v2

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/http/coreapi"
	"github.com/centrifuge/go-centrifuge/nft"
	"github.com/centrifuge/go-centrifuge/utils/httputils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/go-chi/chi"
	"github.com/go-chi/render"
)

const (
	// ErrInvalidTokenID is a sentinel error when token ID is invalid
	ErrInvalidTokenID = errors.Error("Invalid Token ID")

	// ErrInvalidRegistryAddress is a sentinel error when registry address is invalid
	ErrInvalidRegistryAddress = errors.Error("Invalid registry address")
)

// MintNFT mints an NFT.
// @summary Mints an NFT against a document.
// @description Mints an NFT against a document.
// @id mint_nft
// @tags NFTs
// @accept json
// @param authorization header string true "Hex encoded centrifuge ID of the account for the intended API action"
// @param registry_address path string true "NFT registry address in hex"
// @param body body coreapi.MintNFTRequest true "Mint NFT request"
// @produce json
// @Failure 403 {object} httputils.HTTPError
// @Failure 500 {object} httputils.HTTPError
// @Failure 400 {object} httputils.HTTPError
// @success 202 {object} coreapi.MintNFTResponse
// @router /v2/nfts/registries/{registry_address}/mint [post]
func (h handler) MintNFT(w http.ResponseWriter, r *http.Request) {
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

	var req coreapi.MintNFTRequest
	err = json.Unmarshal(data, &req)
	if err != nil {
		code = http.StatusBadRequest
		log.Error(err)
		return
	}

	resp, err := h.srv.MintNFT(ctx, coreapi.ToNFTMintRequest(req, registry))
	if err != nil {
		code = http.StatusBadRequest
		log.Error(err)
		return
	}

	nftResp := coreapi.MintNFTResponse{
		Header:          coreapi.NFTResponseHeader{JobID: resp.JobID},
		RegistryAddress: registry,
		DepositAddress:  req.DepositAddress,
		DocumentID:      req.DocumentID,
		TokenID:         resp.TokenID,
	}
	render.Status(r, http.StatusAccepted)
	render.JSON(w, r, nftResp)
}

// TransferNFT transfers given NFT to provide address.
// @summary Transfers given NFT to provide address.
// @description Transfers given NFT to provide address.
// @id transfer_nft
// @tags NFTs
// @accept json
// @param authorization header string true "Hex encoded centrifuge ID of the account for the intended API action"
// @param registry_address path string true "NFT registry address in hex"
// @param token_id path string true "NFT token ID in hex"
// @param body body coreapi.TransferNFTRequest true "Mint NFT request"
// @produce json
// @Failure 403 {object} httputils.HTTPError
// @Failure 500 {object} httputils.HTTPError
// @Failure 400 {object} httputils.HTTPError
// @success 200 {object} coreapi.TransferNFTResponse
// @router /v2/nfts/registries/{registry_address}/tokens/{token_id}/transfer [post]
func (h handler) TransferNFT(w http.ResponseWriter, r *http.Request) {
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

	var req coreapi.TransferNFTRequest
	err = json.Unmarshal(data, &req)
	if err != nil {
		code = http.StatusBadRequest
		log.Error(err)
		return
	}

	resp, err := h.srv.TransferNFT(ctx, req.To, registry, tokenID)
	if err != nil {
		code = http.StatusBadRequest
		log.Error(err)
		return
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, coreapi.TransferNFTResponse{
		RegistryAddress: registry,
		To:              req.To,
		TokenID:         resp.TokenID,
		Header:          coreapi.NFTResponseHeader{JobID: resp.JobID},
	})
}

// OwnerOfNFT returns the owner of the given NFT.
// @summary Returns the Owner of the given NFT.
// @description Returns the Owner of the given NFT.
// @id owner_of_nft
// @tags NFTs
// @param authorization header string true "Hex encoded centrifuge ID of the account for the intended API action"
// @param token_id path string true "NFT token ID in hex"
// @param registry_address path string true "Registry address in hex"
// @produce json
// @Failure 403 {object} httputils.HTTPError
// @Failure 400 {object} httputils.HTTPError
// @success 200 {object} coreapi.NFTOwnerResponse
// @router /v2/nfts/registries/{registry_address}/tokens/{token_id}/owner [get]
func (h handler) OwnerOfNFT(w http.ResponseWriter, r *http.Request) {
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
	owner, err := h.srv.OwnerOfNFT(registry, tokenID)
	if err != nil {
		code = http.StatusBadRequest
		log.Error(err)
		return
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, coreapi.NFTOwnerResponse{
		TokenID:         tokenID.String(),
		RegistryAddress: registry,
		Owner:           owner,
	})
}
