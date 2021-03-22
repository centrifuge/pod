// +build unit

package testingnfts

import (
	"context"

	"github.com/centrifuge/go-centrifuge/nft"
	"github.com/centrifuge/go-substrate-rpc-client/v2/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/mock"
)

// MockNFTService mocks NFT service
type MockNFTService struct {
	mock.Mock
}

func (m *MockNFTService) MintNFT(ctx context.Context, request nft.MintNFTRequest) (*nft.TokenResponse, error) {
	args := m.Called(ctx, request)
	resp, _ := args.Get(0).(*nft.TokenResponse)
	return resp, args.Error(1)
}

func (m *MockNFTService) MintNFTOnCC(ctx context.Context, request nft.MintNFTOnCCRequest) (*nft.TokenResponse, error) {
	args := m.Called(ctx, request)
	resp, _ := args.Get(0).(*nft.TokenResponse)
	return resp, args.Error(1)
}

func (m *MockNFTService) TransferFrom(ctx context.Context, registry common.Address, to common.Address, tokenID nft.TokenID) (*nft.TokenResponse, error) {
	args := m.Called(ctx)
	resp, _ := args.Get(0).(*nft.TokenResponse)
	return resp, args.Error(1)
}

func (m *MockNFTService) OwnerOf(registry common.Address, tokenID []byte) (owner common.Address, err error) {
	args := m.Called(registry, tokenID)
	resp, _ := args.Get(0).(common.Address)
	return resp, args.Error(1)
}

func (m *MockNFTService) OwnerOfWithRetrial(registry common.Address, tokenID []byte) (owner common.Address, err error) {
	args := m.Called(registry, tokenID)
	resp, _ := args.Get(0).(common.Address)
	return resp, args.Error(1)
}

func (m *MockNFTService) OwnerOfOnCC(registry common.Address, tokenID []byte) (types.AccountID, error) {
	args := m.Called(registry, tokenID)
	acc, _ := args.Get(0).(types.AccountID)
	return acc, args.Error(1)
}

func (m *MockNFTService) TransferNFT(ctx context.Context, registry common.Address, tokenID nft.TokenID,
	to types.AccountID) (*nft.TokenResponse, error) {
	args := m.Called(ctx, registry, tokenID, to)
	tr, _ := args.Get(0).(*nft.TokenResponse)
	return tr, args.Error(1)
}
