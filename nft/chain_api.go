package nft

import (
	"context"

	"github.com/centrifuge/go-centrifuge/centchain"
	"github.com/centrifuge/go-centrifuge/contextutil"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/centrifuge/go-substrate-rpc-client/types"
)

const (
	// ValidateMint is the module call for validate NFT Mint on Centrifuge chain
	ValidateMint = "Nfts.validate_mint"
	// TargetChainID is the target chain where to mint the NFT against - 0 Ethereum
	TargetChainID = 0
)

// API defines set of functions to interact with centrifuge chain
type API interface {
	// ValidateNFT validates the proofs and triggers a bridge event to mint NFT on Ethereum chain.
	ValidateNFT(
		ctx context.Context,
		anchorID [32]byte,
		depositAddress [20]byte,
		proofs []SubstrateProof,
		staticProofs [3][32]byte) (err error)
}

// SubstrateProof holds a single proof value with specific types that goes hand in hand with types on cent chain
type SubstrateProof struct {
	LeafHash     [32]byte
	SortedHashes [][32]byte
}

func toSubstrateProofs(props, values [][]byte, salts [][32]byte, sortedHashes [][][32]byte) (proofs []SubstrateProof) {
	for i := 0; i < len(props); i++ {
		leafHash := utils.MustSliceToByte32(getLeafHash(props[i], values[i], salts[i]))
		proofs = append(proofs, SubstrateProof{
			LeafHash:     leafHash,
			SortedHashes: sortedHashes[i],
		})
	}

	return proofs
}

type api struct {
	api centchain.API
}

func (a api) ValidateNFT(
	ctx context.Context,
	anchorID [32]byte,
	depositAddress [20]byte,
	proofs []SubstrateProof,
	staticProofs [3][32]byte) (err error) {
	acc, err := contextutil.Account(ctx)
	if err != nil {
		return err
	}

	krp, err := acc.GetCentChainAccount().KeyRingPair()
	if err != nil {
		return err
	}

	meta, err := a.api.GetMetadataLatest()
	if err != nil {
		return err
	}

	c, err := types.NewCall(
		meta,
		ValidateMint,
		types.NewHash(anchorID[:]),
		depositAddress,
		proofs,
		staticProofs,
		types.NewU8(TargetChainID))
	if err != nil {
		return err
	}

	return a.api.SubmitAndWatch(ctx, meta, c, krp)
}
