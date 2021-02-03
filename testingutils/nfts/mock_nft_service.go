// +build unit

package testingnfts

import (
	"context"
	"math/big"

	"github.com/centrifuge/go-centrifuge/nft"
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

func (m *MockNFTService) TransferFrom(ctx context.Context, registry common.Address, to common.Address, tokenID nft.TokenID) (*nft.TokenResponse, chan error, error) {
	args := m.Called(ctx)
	resp, _ := args.Get(0).(*nft.TokenResponse)
	done, _ := args.Get(1).(chan error)
	return resp, done, args.Error(2)
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

func (m *MockNFTService) CurrentIndexOfToken(registry common.Address, tokenID []byte) (*big.Int, error) {
	args := m.Called(registry, tokenID)
	resp, _ := args.Get(0).(*big.Int)
	return resp, args.Error(1)
}
