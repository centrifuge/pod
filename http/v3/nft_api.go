package v3

import (
	"encoding/json"
	"io/ioutil"
	"math/big"
	"net/http"
	"strconv"

	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/http/coreapi"
	nftv3 "github.com/centrifuge/go-centrifuge/nft/v3"
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
// @param class_id path string true "NFT class ID"
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

	classIDParam, err := strconv.Atoi(chi.URLParam(r, coreapi.ClassIDParam))

	if err != nil {
		code = http.StatusBadRequest
		err = ErrInvalidClassID
		h.log.Error(err)
		return
	}

	classID := types.U64(classIDParam)

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

	nftResp := coreapi.MintNFTV3Response{
		Header: coreapi.NFTResponseHeader{
			JobID: res.JobID,
		},
		DocumentID:     req.DocumentID,
		ClassID:        classID,
		InstanceID:     res.InstanceID.String(),
		Owner:          req.Owner,
		Metadata:       req.Metadata,
		FreezeMetadata: req.FreezeMetadata,
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
// @param class_id path string true "NFT class ID"
// @param instance_id path string true "NFT instance ID"
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

	classIDParam, err := strconv.Atoi(chi.URLParam(r, coreapi.ClassIDParam))

	if err != nil {
		code = http.StatusBadRequest
		err = ErrInvalidClassID
		h.log.Error(err)
		return
	}

	classID := types.U64(classIDParam)

	instanceIDParam := chi.URLParam(r, coreapi.InstanceIDParam)

	b := new(big.Int)
	i, ok := b.SetString(instanceIDParam, 10)

	if !ok {
		code = http.StatusBadRequest
		err = ErrInvalidInstanceID
		h.log.Error(err)
		return
	}

	instanceID := types.NewU128(*i)

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
		ClassID:    classID,
		InstanceID: instanceID.String(),
		Owner:      owner,
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, ownerOfResp)
}

// CreateNFTClass creates an NFT class on the Centrifuge chain.
// @summary Creates a specific NFT class.
// @description Creates a specific NFT class
// @id create_nft_class
// @tags NFTs
// @accept json
// @param authorization header string true "Hex encoded centrifuge ID of the account for the intended API action"
// @param body body coreapi.CreateNFTClassV3Request true "Mint NFT request V3"
// @produce json
// @Failure 403 {object} httputils.HTTPError
// @Failure 500 {object} httputils.HTTPError
// @Failure 400 {object} httputils.HTTPError
// @success 202 {object} coreapi.CreateNFTClassV3Response
// @router /v3/nfts/classes [post]
func (h *handler) CreateNFTClass(w http.ResponseWriter, r *http.Request) {
	var err error
	var code int
	defer httputils.RespondIfError(&code, &err, w, r)

	requestBody, err := ioutil.ReadAll(r.Body)

	if err != nil {
		code = http.StatusInternalServerError
		h.log.Error(err)
		return
	}

	var req coreapi.CreateNFTClassV3Request

	if err := json.Unmarshal(requestBody, &req); err != nil {
		code = http.StatusBadRequest
		h.log.Error(err)
		return
	}

	ctx := r.Context()

	res, err := h.srv.CreateNFTClass(ctx, &nftv3.CreateNFTClassRequest{ClassID: req.ClassID})

	if err != nil {
		code = http.StatusBadRequest
		h.log.Error(err)
		return
	}

	nftResp := coreapi.CreateNFTClassV3Response{
		Header: coreapi.NFTResponseHeader{
			JobID: res.JobID,
		},
		ClassID: req.ClassID,
	}

	render.Status(r, http.StatusAccepted)
	render.JSON(w, r, nftResp)
}

// MetadataOfNFT returns the metadata of an NFT instance.
// @summary Returns the metadata of an NFT instance.
// @description Returns the metadata of an NFT instance.
// @id metadata_of_nft
// @tags NFTs
// @param authorization header string true "Hex encoded centrifuge ID of the account for the intended API action"
// @param class_id path string true "NFT class ID"
// @param instance_id path string true "NFT instance ID"
// @produce json
// @Failure 403 {object} httputils.HTTPError
// @Failure 404 {object} httputils.HTTPError
// @Failure 500 {object} httputils.HTTPError
// @Failure 400 {object} httputils.HTTPError
// @success 200 {object} coreapi.InstanceMetadataOfNFTV3Response
// @router /v3/nfts/classes/{class_id}/instances/{instance_id}/metadata [get]
func (h *handler) MetadataOfNFT(w http.ResponseWriter, r *http.Request) {
	var err error
	var code int
	defer httputils.RespondIfError(&code, &err, w, r)

	classIDParam, err := strconv.Atoi(chi.URLParam(r, coreapi.ClassIDParam))

	if err != nil {
		code = http.StatusBadRequest
		err = ErrInvalidClassID
		h.log.Error(err)
		return
	}

	classID := types.U64(classIDParam)

	instanceIDParam := chi.URLParam(r, coreapi.InstanceIDParam)

	b := new(big.Int)
	i, ok := b.SetString(instanceIDParam, 10)

	if !ok {
		code = http.StatusBadRequest
		err = ErrInvalidInstanceID
		h.log.Error(err)
		return
	}

	instanceID := types.NewU128(*i)

	ctx := r.Context()

	res, err := h.srv.InstanceMetadataOfNFT(
		ctx,
		&nftv3.InstanceMetadataOf{
			ClassID:    classID,
			InstanceID: instanceID,
		},
	)

	if err != nil {
		code = http.StatusBadRequest

		if errors.IsOfType(err, nftv3.ErrInstanceMetadataNotFound) {
			code = http.StatusNotFound
		}

		h.log.Error(err)
		return
	}

	instanceMetadataOfResp := coreapi.InstanceMetadataOfNFTV3Response{
		Deposit:  res.Deposit.String(),
		Data:     string(res.Data),
		IsFrozen: res.IsFrozen,
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, instanceMetadataOfResp)
}
