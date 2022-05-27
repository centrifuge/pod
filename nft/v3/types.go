package v3

import (
	"encoding/gob"

	"github.com/centrifuge/go-centrifuge/documents"

	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
)

func init() {
	gob.Register(types.U64(0))
	gob.Register(types.U128{})
	gob.Register(MintNFTRequest{})
}

// OwnerOfRequest is the request object for the retrieval of the owner of an NFT on Centrifuge chain.
type OwnerOfRequest struct {
	ClassID    types.U64
	InstanceID types.U128
}

// OwnerOfResponse is the response object for a OwnerOfRequest, it holds the AccountID of the owner of an NFT.
type OwnerOfResponse struct {
	ClassID    types.U64
	InstanceID types.U128
	AccountID  types.AccountID
}

// MintNFTRequest is the request object for minting an NFT on Centrifuge chain.
type MintNFTRequest struct {
	DocumentID     []byte
	ClassID        types.U64
	Owner          types.AccountID // substrate account ID
	DocAttributes  []documents.AttrKey
	FreezeMetadata bool
}

// MintNFTResponse is the response object for a MintNFTRequest, it holds the job ID and instance ID of the NFT.
type MintNFTResponse struct {
	JobID      string
	InstanceID types.U128
}

// CreateNFTClassRequest is the response object for creating an NFT class on Centrifuge chain.
type CreateNFTClassRequest struct {
	// TODO(cdamian): Add more fields such as admin?
	ClassID types.U64
}

// CreateNFTClassResponse is the response object for a CreateNFTClassRequest, it holds the job ID and the newly created class ID.
type CreateNFTClassResponse struct {
	JobID   string
	ClassID types.U64
}

// InstanceMetadataOfRequest is the request object for retrieving the metadata of an NFT instance.
type InstanceMetadataOfRequest struct {
	ClassID    types.U64
	InstanceID types.U128
}

// NFTMetadata is the struct of the NFT metadata that is stored in IPFS.
type NFTMetadata struct {
	DocID         []byte        `json:"doc_id"`
	DocVersion    []byte        `json:"doc_version"`
	DocAttributes DocAttributes `json:"doc_attributes"`
}

// DocAttributes is a map of document attributes to their respective values.
type DocAttributes map[string]string
