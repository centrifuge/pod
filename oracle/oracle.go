package oracle

import (
	"context"

	"github.com/ethereum/go-ethereum/common"
)

const (
	// TokenIDLength is the length of an NFT token ID
	TokenIDLength = 32
)

type updateNFTOracleRequest struct {
	TokenID [TokenIDLength]byte
	// this is the unique identifier of the NFT Oracle contract
	OracleFingerprint []byte
	Result            []byte
	OracleAddress     common.Address
}

type Service interface {
	// Interacts with the NFT Oracle contract to update it
	UpdateNFTOracle(ctx context.Context, request updateNFTOracleRequest)
}

// UpdateResponse holds the job ID
type UpdateResponse struct {
	JobID string
}
