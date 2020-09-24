package oracle

import (
	"context"

	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/nft"
	"github.com/ethereum/go-ethereum/common"
)

// PushAttributeToOracleRequest request holds the required data to push value to oracle
type PushAttributeToOracleRequest struct {
	TokenID       nft.TokenID       `json:"token_id"`
	AttributeKey  documents.AttrKey `json:"attribute_key"`
	OracleAddress common.Address    `json:"oracle_address"`
}

// Service defines the functions to Oracle
type Service interface {
	// PushAttributeToOracle pushes a given
	PushAttributeToOracle(ctx context.Context, docID []byte, request PushAttributeToOracleRequest) (*PushToOracleResponse, error)
}

// UpdateResponse holds the job ID
type PushToOracleResponse struct {
	JobID string `json:"job_id"`
	PushAttributeToOracleRequest
}
