package nft

import (
	"context"
	"fmt"

	"github.com/centrifuge/go-centrifuge/centchain"
	"github.com/centrifuge/go-centrifuge/contextutil"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/centrifuge/go-substrate-rpc-client/v2/types"
	"github.com/ethereum/go-ethereum/common"
)

const (
	// ValidateMint is the module call for validate NFT Mint on Centrifuge chain
	ValidateMint = "Nfts.validate_mint"
	// TargetChainID is the target chain where to mint the NFT against - 0 Ethereum
	TargetChainID = 0
)

// API defines set of functions to interact with centrifuge chain
type API interface {
	// ValidateNFT validates the Proofs and triggers a bridge event to mint NFT on Ethereum chain.
	ValidateNFT(
		ctx context.Context,
		anchorID [32]byte,
		depositAddress [20]byte,
		proofs []SubstrateProof,
		staticProofs [3][32]byte) (err error)

	// CreateRegistry creates a new nft registry on centrifuge chain
	CreateRegistry(ctx context.Context, info RegistryInfo) (registryID common.Address, err error)

	// MintNFT sends an extrinsic to mint nft in given registry on cent chain
	MintNFT(ctx context.Context, owner types.AccountID, registry types.H160, tokenID types.U256, assetInfo AssetInfo, mintInfo MintInfo) (info centchain.ExtrinsicInfo, err error)
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

func toNFTOnCCProofs(props, values [][]byte, salts [][32]byte, sortedHashes [][][32]byte) (proofs []Proof) {
	for i := 0; i < len(props); i++ {
		proofs = append(proofs, Proof{
			Value:    values[i],
			Property: props[i],
			Salt:     salts[i],
			Hashes:   sortedHashes[i],
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

	_, err = a.api.SubmitAndWatch(ctx, meta, c, krp)
	return err
}

// RegistryInfo is used as parameter to create registry on cent chain
type RegistryInfo struct {
	OwnerCanBurn bool
	Fields       [][]byte
}

// CreateRegistry creates a new NFT registry on centrifuge chain
func (a api) CreateRegistry(ctx context.Context, info RegistryInfo) (registryID common.Address, err error) {
	acc, err := contextutil.Account(ctx)
	if err != nil {
		return registryID, err
	}

	krp, err := acc.GetCentChainAccount().KeyRingPair()
	if err != nil {
		return registryID, err
	}

	meta, err := a.api.GetMetadataLatest()
	if err != nil {
		return registryID, err
	}

	call, err := types.NewCall(meta, "Registry.create_registry", info)
	if err != nil {
		return registryID, fmt.Errorf("failed to create extrinsic: %w", err)
	}

	extInfo, err := a.api.SubmitAndWatch(ctx, meta, call, krp)
	if err != nil {
		return registryID, fmt.Errorf("failed to create registry: %w", err)
	}

	events, err := extInfo.Events(meta)
	if err != nil {
		return registryID, fmt.Errorf("failed to decode events: %w", err)
	}

	for _, e := range events.Registry_RegistryCreated {
		if !e.Phase.IsApplyExtrinsic {
			continue
		}

		if uint(e.Phase.AsApplyExtrinsic) == extInfo.Index {
			return common.Address(e.RegistryID), nil
		}
	}

	return registryID, errors.New("failed to create registry")
}

// AssetInfo contains metadata of an nft
type AssetInfo struct {
	Metadata []byte
}

// Proof is single nft proof
type Proof struct {
	Value    []byte
	Property []byte
	Salt     [32]byte
	Hashes   [][32]byte
}

// MintInfo has Proofs to be validated to mint NFT
type MintInfo struct {
	AnchorID     [32]byte
	StaticHashes [3][32]byte
	Proofs       []Proof
}

// MintNFT sends an extrinsic to mint nft in given registry on cent chain
func (a api) MintNFT(
	ctx context.Context,
	owner types.AccountID,
	registry types.H160, tokenID types.U256,
	assetInfo AssetInfo,
	mintInfo MintInfo) (info centchain.ExtrinsicInfo, err error) {
	acc, err := contextutil.Account(ctx)
	if err != nil {
		return info, err
	}

	krp, err := acc.GetCentChainAccount().KeyRingPair()
	if err != nil {
		return info, err
	}

	meta, err := a.api.GetMetadataLatest()
	if err != nil {
		return info, err
	}

	call, err := types.NewCall(meta, "Registry.mint", owner, registry, tokenID, assetInfo, mintInfo)
	if err != nil {
		return info, fmt.Errorf("failed to create extrinsic: %w", err)
	}

	info, err = a.api.SubmitAndWatch(ctx, meta, call, krp)
	if err != nil {
		return info, fmt.Errorf("failed to mint nft: %w", err)
	}

	return info, nil
}
