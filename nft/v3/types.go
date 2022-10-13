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

// CreateNFTCollectionResponse is the response object for a CreateNFTCollectionRequest, it holds the job ID and the newly created class ID.
type CreateNFTCollectionResponse struct {
	JobID        string
	CollectionID types.U64
}
