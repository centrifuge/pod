package v3

import (
	"encoding/json"
	"io/ioutil"
	"math/big"
	"net/http"
	"strconv"

	"github.com/centrifuge/pod/utils/byteutils"

	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/centrifuge/pod/errors"
	"github.com/centrifuge/pod/http/coreapi"
	nftv3 "github.com/centrifuge/pod/nft/v3"
	"github.com/centrifuge/pod/utils/httputils"
	"github.com/go-chi/chi"
	"github.com/go-chi/render"
)

const (
	// ErrInvalidCollectionID is a sentinel error when the collection ID is invalid.
	ErrInvalidCollectionID = errors.Error("invalid collection ID")

	// ErrInvalidItemID is a sentinel error when the item ID is invalid.
	ErrInvalidItemID = errors.Error("invalid item ID")
)

// CommitAndMintNFT commits a pending document and mints an NFT on the Centrifuge chain.
// @summary commits a pending document and mints an NFT on the Centrifuge chain.
// @description commits a pending document and mints an NFT on the Centrifuge chain.
// @id commit_and_mint_nft
// @tags NFTs
// @accept json
// @param authorization header string true "Bearer <JW3T token>"
// @param collection_id path string true "NFT collection ID"
// @param body body coreapi.MintNFTV3Request true "Mint NFT request V3"
// @produce json
// @Failure 403 {object} httputils.HTTPError
// @Failure 500 {object} httputils.HTTPError
// @Failure 400 {object} httputils.HTTPError
// @success 202 {object} coreapi.MintNFTV3Response
// @router /v3/nfts/collections/{collection_id}/commit_and_mint [post]
func (h *handler) CommitAndMintNFT(w http.ResponseWriter, r *http.Request) {
	h.mintNFT(w, r, true)
}

// MintNFT mints an NFT on the Centrifuge chain.
// @summary Mints an NFT for a specified document.
// @description Mints an NFT for a specified document.
// @id mint_nft
// @tags NFTs
// @accept json
// @param authorization header string true "Bearer <JW3T token>"
// @param collection_id path string true "NFT collection ID"
// @param body body coreapi.MintNFTV3Request true "Mint NFT request V3"
// @produce json
// @Failure 403 {object} httputils.HTTPError
// @Failure 500 {object} httputils.HTTPError
// @Failure 400 {object} httputils.HTTPError
// @success 202 {object} coreapi.MintNFTV3Response
// @router /v3/nfts/collections/{collection_id}/mint [post]
func (h *handler) MintNFT(w http.ResponseWriter, r *http.Request) {
	h.mintNFT(w, r, false)
}

func (h *handler) mintNFT(w http.ResponseWriter, r *http.Request, documentPending bool) {
	var err error
	var code int
	defer httputils.RespondIfError(&code, &err, w, r)

	collectionIDParam, err := strconv.Atoi(chi.URLParam(r, coreapi.CollectionIDParam))

	if err != nil {
		code = http.StatusBadRequest
		err = ErrInvalidCollectionID
		h.log.Error(err)
		return
	}

	collectionID := types.U64(collectionIDParam)

	requestBody, err := ioutil.ReadAll(r.Body)

	if err != nil {
		code = http.StatusInternalServerError
		h.log.Error(err)
		err = coreapi.ErrRequestBodyRead
		return
	}

	var req coreapi.MintNFTV3Request

	if err = json.Unmarshal(requestBody, &req); err != nil {
		code = http.StatusBadRequest
		h.log.Error(err)
		err = coreapi.ErrRequestPayloadJSONDecode
		return
	}

	ctx := r.Context()

	res, err := h.srv.MintNFT(
		ctx,
		coreapi.ToNFTMintRequestV3(req, collectionID),
		documentPending,
	)

	if err != nil {
		code = http.StatusBadRequest
		h.log.Error(err)
		return
	}

	nftResp := coreapi.MintNFTV3Response{
		Header: coreapi.NFTResponseHeader{
			JobID: res.JobID,
		},
		DocumentID:   req.DocumentID,
		CollectionID: collectionID,
		ItemID:       res.ItemID.String(),
		Owner:        req.Owner,
		IPFSMetadata: req.IPFSMetadata,
	}

	render.Status(r, http.StatusAccepted)
	render.JSON(w, r, nftResp)
}

// GetNFTOwner returns the owner of an NFT on Centrifuge chain.
// @summary Returns the owner of an NFT.
// @description Returns the owner of an NFT.
// @id owner_of_nft
// @tags NFTs
// @param authorization header string true "Bearer <JW3T token>"
// @param collection_id path string true "NFT collection ID"
// @param item_id path string true "NFT item ID"
// @produce json
// @Failure 403 {object} httputils.HTTPError
// @Failure 404 {object} httputils.HTTPError
// @Failure 500 {object} httputils.HTTPError
// @Failure 400 {object} httputils.HTTPError
// @success 200 {object} coreapi.GetNFTOwnerV3Response
// @router /v3/nfts/collections/{collection_id}/items/{item_id}/owner [get]
func (h *handler) GetNFTOwner(w http.ResponseWriter, r *http.Request) {
	var err error
	var code int
	defer httputils.RespondIfError(&code, &err, w, r)

	collectionIDParam, err := strconv.Atoi(chi.URLParam(r, coreapi.CollectionIDParam))

	if err != nil {
		code = http.StatusBadRequest
		h.log.Error(err)
		err = ErrInvalidCollectionID
		return
	}

	collectionID := types.U64(collectionIDParam)

	itemIDParam := chi.URLParam(r, coreapi.ItemIDParam)

	b := new(big.Int)
	i, ok := b.SetString(itemIDParam, 10)

	if !ok {
		code = http.StatusBadRequest
		h.log.Error(err)
		err = ErrInvalidItemID
		return
	}

	itemID := types.NewU128(*i)

	owner, err := h.srv.GetNFTOwner(collectionID, itemID)

	if err != nil {
		code = http.StatusBadRequest

		if errors.IsOfType(err, nftv3.ErrOwnerNotFound) {
			code = http.StatusNotFound
		}

		h.log.Error(err)
		return
	}

	ownerOfResp := coreapi.GetNFTOwnerV3Response{
		CollectionID: collectionID,
		ItemID:       itemID.String(),
		Owner:        owner,
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, ownerOfResp)
}

// CreateNFTCollection creates an NFT collection on the Centrifuge chain.
// @summary Creates a specific NFT collection.
// @description Creates a specific NFT collection
// @id create_nft_collection
// @tags NFTs
// @accept json
// @param authorization header string true "Bearer <JW3T token>"
// @param body body coreapi.CreateNFTCollectionV3Request true "Create NFT collection request V3"
// @produce json
// @Failure 403 {object} httputils.HTTPError
// @Failure 500 {object} httputils.HTTPError
// @Failure 400 {object} httputils.HTTPError
// @success 202 {object} coreapi.CreateNFTCollectionV3Response
// @router /v3/nfts/collections [post]
func (h *handler) CreateNFTCollection(w http.ResponseWriter, r *http.Request) {
	var err error
	var code int
	defer httputils.RespondIfError(&code, &err, w, r)

	requestBody, err := ioutil.ReadAll(r.Body)

	if err != nil {
		code = http.StatusInternalServerError
		h.log.Error(err)
		return
	}

	var req coreapi.CreateNFTCollectionV3Request

	if err = json.Unmarshal(requestBody, &req); err != nil {
		code = http.StatusBadRequest
		h.log.Error(err)
		return
	}

	ctx := r.Context()

	res, err := h.srv.CreateNFTCollection(ctx, req.CollectionID)

	if err != nil {
		code = http.StatusBadRequest
		h.log.Error(err)
		return
	}

	nftResp := coreapi.CreateNFTCollectionV3Response{
		Header: coreapi.NFTResponseHeader{
			JobID: res.JobID,
		},
		CollectionID: req.CollectionID,
	}

	render.Status(r, http.StatusAccepted)
	render.JSON(w, r, nftResp)
}

// MetadataOfNFT returns the metadata of an NFT item.
// @summary Returns the metadata of an NFT item.
// @description Returns the metadata of an NFT item.
// @id metadata_of_nft
// @tags NFTs
// @param authorization header string true "Bearer <JW3T token>"
// @param collection_id path string true "NFT collection ID"
// @param item_id path string true "NFT item ID"
// @produce json
// @Failure 403 {object} httputils.HTTPError
// @Failure 404 {object} httputils.HTTPError
// @Failure 500 {object} httputils.HTTPError
// @Failure 400 {object} httputils.HTTPError
// @success 200 {object} coreapi.ItemMetadataOfNFTV3Response
// @router /v3/nfts/collections/{collection_id}/items/{item_id}/metadata [get]
func (h *handler) MetadataOfNFT(w http.ResponseWriter, r *http.Request) {
	var err error
	var code int
	defer httputils.RespondIfError(&code, &err, w, r)

	collectionIDParam, err := strconv.Atoi(chi.URLParam(r, coreapi.CollectionIDParam))

	if err != nil {
		code = http.StatusBadRequest
		h.log.Error(err)
		err = ErrInvalidCollectionID
		return
	}

	collectionID := types.U64(collectionIDParam)

	itemIDParam := chi.URLParam(r, coreapi.ItemIDParam)

	b := new(big.Int)
	i, ok := b.SetString(itemIDParam, 10)

	if !ok {
		code = http.StatusBadRequest
		h.log.Error(err)
		err = ErrInvalidItemID
		return
	}

	itemID := types.NewU128(*i)

	res, err := h.srv.GetItemMetadata(collectionID, itemID)

	if err != nil {
		code = http.StatusBadRequest

		if errors.IsOfType(err, nftv3.ErrItemMetadataNotFound) {
			code = http.StatusNotFound
		}

		h.log.Error(err)
		return
	}

	itemMetadataResp := coreapi.ItemMetadataOfNFTV3Response{
		Deposit:  res.Deposit.String(),
		Data:     byteutils.HexBytes(res.Data),
		IsFrozen: res.IsFrozen,
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, itemMetadataResp)
}

// AttributeOfNFT returns the attribute of an NFT item.
// @summary Returns the attribute of an NFT item.
// @description Returns the attribute of an NFT item.
// @id attribute_of_nft
// @tags NFTs
// @param authorization header string true "Bearer <JW3T token>"
// @param collection_id path string true "NFT collection ID"
// @param item_id path string true "NFT item ID"
// @param attribute_name path string true "NFT attribute name"
// @produce json
// @Failure 403 {object} httputils.HTTPError
// @Failure 404 {object} httputils.HTTPError
// @Failure 500 {object} httputils.HTTPError
// @Failure 400 {object} httputils.HTTPError
// @success 200 {object} coreapi.ItemAttributeOfNFTV3Response
// @router /v3/nfts/collections/{collection_id}/items/{item_id}/attribute/{attribute_name} [get]
func (h *handler) AttributeOfNFT(w http.ResponseWriter, r *http.Request) {
	var err error
	var code int
	defer httputils.RespondIfError(&code, &err, w, r)

	collectionIDParam, err := strconv.Atoi(chi.URLParam(r, coreapi.CollectionIDParam))

	if err != nil {
		code = http.StatusBadRequest
		h.log.Error(err)
		err = ErrInvalidCollectionID
		return
	}

	collectionID := types.U64(collectionIDParam)

	itemIDParam := chi.URLParam(r, coreapi.ItemIDParam)

	b := new(big.Int)
	i, ok := b.SetString(itemIDParam, 10)

	if !ok {
		code = http.StatusBadRequest
		h.log.Error(err)
		err = ErrInvalidItemID
		return
	}

	itemID := types.NewU128(*i)

	res, err := h.srv.GetItemAttribute(collectionID, itemID, chi.URLParam(r, coreapi.AttributeNameParam))

	if err != nil {
		code = http.StatusBadRequest

		if errors.IsOfType(err, nftv3.ErrItemAttributeNotFound) {
			code = http.StatusNotFound
		}

		h.log.Error(err)
		return
	}

	itemAttributeResponse := coreapi.ItemAttributeOfNFTV3Response{
		Value: res,
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, itemAttributeResponse)
}
