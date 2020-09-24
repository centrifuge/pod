package oracle

import (
	"context"

	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/nft"
	"github.com/ethereum/go-ethereum/common"
)

// PushAttributeToOracleRequest request holds the required data to push value to oracle
type PushAttributeToOracleRequest struct {
	TokenID       nft.TokenID       `json:"token_id" swaggertype:"primitive,string"`       // hex value of the NFT token
	AttributeKey  documents.AttrKey `json:"attribute_key" swaggertype:"primitive,string"`  // hex value of the Attribute key
	OracleAddress common.Address    `json:"oracle_address" swaggertype:"primitive,string"` // hex value of the Oracle address
}

// Service defines the functions to Oracle
type Service interface {
	// PushAttributeToOracle pushes a given
	PushAttributeToOracle(ctx context.Context, docID []byte, request PushAttributeToOracleRequest) (*PushToOracleResponse, error)
}

// PushToOracleResponse holds the job ID along with original request.
type PushToOracleResponse struct {
	JobID string `json:"job_id"`
	PushAttributeToOracleRequest
}
