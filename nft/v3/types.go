package v3

import (
	"encoding/gob"

	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
)

func init() {
	gob.Register(types.U64(0))
	gob.Register(types.U128{})
	gob.Register(&MintNFTRequest{})
}

// OwnerOfRequest is the request object for the retrieval of the owner of an NFT on Centrifuge chain.
type OwnerOfRequest struct {
	CollectionID types.U64
	ItemID       types.U128
}

// OwnerOfResponse is the response object for a OwnerOfRequest, it holds the AccountID of the owner of an NFT.
type OwnerOfResponse struct {
	CollectionID types.U64
	ItemID       types.U128
	AccountID    *types.AccountID
}

// MintNFTRequest is the request object for minting an NFT on Centrifuge chain.
type MintNFTRequest struct {
	DocumentID      []byte
	CollectionID    types.U64
	Owner           *types.AccountID // substrate account ID
	IPFSMetadata    IPFSMetadata
	GrantReadAccess bool
}

type IPFSMetadata struct {
	Name                  string   `json:"name"`
	Description           string   `json:"description,omitempty"`
	Image                 string   `json:"image,omitempty"`
	DocumentAttributeKeys []string `json:"document_attribute_keys"`
}

// MintNFTResponse is the response object for a MintNFTRequest, it holds the job ID and instance ID of the NFT.
type MintNFTResponse struct {
	JobID  string
	ItemID types.U128
}

// CreateNFTCollectionRequest is the response object for creating an NFT class on Centrifuge chain.
type CreateNFTCollectionRequest struct {
	CollectionID types.U64
}

// CreateNFTCollectionResponse is the response object for a CreateNFTCollectionRequest, it holds the job ID and the newly created class ID.
type CreateNFTCollectionResponse struct {
	JobID        string
	CollectionID types.U64
}

// GetItemMetadataRequest is the request object for retrieving the metadata of an NFT item.
type GetItemMetadataRequest struct {
	CollectionID types.U64
	ItemID       types.U128
}

// GetItemAttributeRequest is the request object for retrieve the attribute of an NFT item.
type GetItemAttributeRequest struct {
	CollectionID types.U64
	ItemID       types.U128
	Key          string
}

// NFTMetadata is the struct of the NFT metadata that is stored in IPFS.
type NFTMetadata struct {
	Name        string            `json:"name"`
	Description string            `json:"description,omitempty"`
	Image       string            `json:"image,omitempty"`
	Properties  map[string]string `json:"properties"`
}
