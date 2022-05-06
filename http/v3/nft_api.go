package v3

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	nftv3 "github.com/centrifuge/go-centrifuge/nft/v3"

	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/http/coreapi"
	"github.com/centrifuge/go-centrifuge/utils/httputils"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/go-chi/chi"
	"github.com/go-chi/render"
)

const (
	// ErrInvalidClassID is a sentinel error when the class ID is invalid.
	ErrInvalidClassID = errors.Error("Invalid class ID")

	// ErrInvalidInstanceID is a sentinel error when the instance ID is invalid.
	ErrInvalidInstanceID = errors.Error("Invalid instance ID")
)

// MintNFT mints an NFT on the Centrifuge chain.
// @summary Mints an NFT for a specified document.
// @description Mints an NFT for a specified document.
// @id mint_nft
// @tags NFTs
// @accept json
// @param authorization header string true "Hex encoded centrifuge ID of the account for the intended API action"
// @param class_id path string true "Hex encoded NFT class ID"
// @param body body coreapi.MintNFTV3Request true "Mint NFT request V3"
// @produce json
// @Failure 403 {object} httputils.HTTPError
// @Failure 500 {object} httputils.HTTPError
// @Failure 400 {object} httputils.HTTPError
// @success 202 {object} coreapi.MintNFTV3Response
// @router /v3/nfts/classes/{class_id}/mint [post]
func (h *handler) MintNFT(w http.ResponseWriter, r *http.Request) {
	var err error
	var code int
	defer httputils.RespondIfError(&code, &err, w, r)

	classIDParam := chi.URLParam(r, coreapi.ClassIDParam)

	var classID types.U64

	err = types.DecodeFromHexString(classIDParam, &classID)

	if err != nil {
		code = http.StatusBadRequest
		h.log.Error(err)
		return
	}

	requestBody, err := ioutil.ReadAll(r.Body)

	if err != nil {
		code = http.StatusInternalServerError
		h.log.Error(err)
		return
	}

	var req coreapi.MintNFTV3Request

	if err := json.Unmarshal(requestBody, &req); err != nil {
		code = http.StatusBadRequest
		h.log.Error(err)
		return
	}

	ctx := r.Context()

	res, err := h.srv.MintNFT(ctx, coreapi.ToNFTMintRequestV3(req, classID))

	if err != nil {
		code = http.StatusBadRequest
		h.log.Error(err)
		return
	}

	instanceIDHex, err := types.EncodeToHexString(res.InstanceID)

	if err != nil {
		code = http.StatusInternalServerError
		h.log.Error(err)
		return
	}

	nftResp := coreapi.MintNFTV3Response{
		Header: coreapi.NFTResponseHeader{
			JobID: res.JobID,
		},
		DocumentID: req.DocumentID,
		ClassID:    classIDParam,
		InstanceID: instanceIDHex,
		Owner:      req.Owner,
	}

	render.Status(r, http.StatusAccepted)
	render.JSON(w, r, nftResp)
}

// OwnerOfNFT returns the owner of an NFT on Centrifuge chain.
// @summary Returns the owner of an NFT.
// @description Returns the owner of an NFT.
// @id owner_of_nft
// @tags NFTs
// @param authorization header string true "Hex encoded centrifuge ID of the account for the intended API action"
// @param class_id path string true "Hex encoded NFT class ID"
// @param instance_id path string true "Hex encoded NFT instance ID"
// @produce json
// @Failure 403 {object} httputils.HTTPError
// @Failure 404 {object} httputils.HTTPError
// @Failure 500 {object} httputils.HTTPError
// @Failure 400 {object} httputils.HTTPError
// @success 200 {object} coreapi.OwnerOfNFTV3Response
// @router /v3/nfts/classes/{class_id}/instances/{instance_id}/owner [get]
func (h *handler) OwnerOfNFT(w http.ResponseWriter, r *http.Request) {
	var err error
	var code int
	defer httputils.RespondIfError(&code, &err, w, r)

	classIDParam := chi.URLParam(r, coreapi.ClassIDParam)

	var classID types.U64

	err = types.DecodeFromHexString(classIDParam, &classID)

	if err != nil {
		code = http.StatusBadRequest
		err = ErrInvalidClassID
		h.log.Error(err)
		return
	}

	instanceIDParam := chi.URLParam(r, coreapi.InstanceIDParam)

	var instanceID types.U128

	err = types.DecodeFromHexString(instanceIDParam, &instanceID)

	if err != nil {
		code = http.StatusBadRequest
		err = ErrInvalidInstanceID
		h.log.Error(err)
		return
	}

	ctx := r.Context()

	res, err := h.srv.OwnerOfNFT(
		ctx,
		&nftv3.OwnerOfRequest{
			ClassID:    classID,
			InstanceID: instanceID,
		},
	)

	if err != nil {
		code = http.StatusBadRequest

		if errors.IsOfType(err, nftv3.ErrInstanceDetailsNotFound) {
			code = http.StatusNotFound
		}

		h.log.Error(err)
		return
	}

	owner, err := types.EncodeToBytes(res.AccountID)

	if err != nil {
		code = http.StatusInternalServerError
		h.log.Error(err)
		return
	}

	ownerOfResp := coreapi.OwnerOfNFTV3Response{
		ClassID:    classIDParam,
		InstanceID: instanceIDParam,
		Owner:      owner,
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, ownerOfResp)
}
