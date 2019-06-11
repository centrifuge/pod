// +build unit

package nft

import (
	"context"
	"math/big"
	"testing"

	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/config/configstore"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/protobufs/gen/go/nft"
	"github.com/centrifuge/go-centrifuge/testingutils/config"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockInvoiceUnpaid struct {
	mock.Mock
}

func (m *mockInvoiceUnpaid) MintNFT(ctx context.Context, request MintNFTRequest) (*TokenResponse, chan bool, error) {
	args := m.Called(ctx, request)
	resp, _ := args.Get(0).(*TokenResponse)
	return resp, nil, args.Error(1)
}

func (m *mockInvoiceUnpaid) GetRequiredInvoiceUnpaidProofFields(ctx context.Context) ([]string, error) {
	args := m.Called(ctx)
	resp, _ := args.Get(0).([]string)
	return resp, args.Error(1)
}

func (m *mockInvoiceUnpaid) TransferFrom(ctx context.Context, registry common.Address, to common.Address, tokenID TokenID) (*TokenResponse, chan bool, error) {
	args := m.Called(ctx)
	resp, _ := args.Get(0).(*TokenResponse)
	return resp, nil, args.Error(1)
}

func (m *mockInvoiceUnpaid) OwnerOf(registry common.Address, tokenID []byte) (owner common.Address, err error) {
	args := m.Called(registry, tokenID)
	resp, _ := args.Get(0).(common.Address)
	return resp, args.Error(1)
}

func TestPaymentObligationNFTMint_success(t *testing.T) {
	mockService := &mockInvoiceUnpaid{}
	mockConfigStore := mockmockConfigStore()
	tokID := big.NewInt(1)
	nftResponse := &TokenResponse{TokenID: tokID.String()}
	nftReq := &nftpb.NFTMintInvoiceUnpaidRequest{
		Identifier:     "0x1234567890",
		DepositAddress: "0xf72855759a39fb75fc7341139f5d7a3974d4da08",
	}

	// error no account header
	handler := grpcHandler{mockConfigStore, mockService}
	nftMintResponse, err := handler.MintInvoiceUnpaidNFT(context.Background(), nftReq)
	mockService.AssertExpectations(t)
	assert.Error(t, err)
	assert.Nil(t, nftMintResponse)

	// error generate proofs
	mockService.On("GetRequiredInvoiceUnpaidProofFields", mock.Anything).Return(nil, errors.New("fail")).Once()
	handler = grpcHandler{mockConfigStore, mockService}
	nftMintResponse, err = handler.MintInvoiceUnpaidNFT(testingconfig.HandlerContext(mockConfigStore), nftReq)
	mockService.AssertExpectations(t)
	assert.Error(t, err)
	assert.Nil(t, nftMintResponse)

	// error get config
	mockService.On("GetRequiredInvoiceUnpaidProofFields", mock.Anything).Return([]string{"proof1", "proof2"}, nil).Once()
	mockConfigStore.On("GetConfig").Return(cfg, errors.New("fail")).Once()
	handler = grpcHandler{mockConfigStore, mockService}
	nftMintResponse, err = handler.MintInvoiceUnpaidNFT(testingconfig.HandlerContext(mockConfigStore), nftReq)
	mockService.AssertExpectations(t)
	assert.Error(t, err)
	assert.Nil(t, nftMintResponse)

	// success assertions
	mockService.On("MintNFT", mock.Anything, mock.Anything).Return(nftResponse, nil).Once()
	mockService.On("GetRequiredInvoiceUnpaidProofFields", mock.Anything).Return([]string{"proof1", "proof2"}, nil).Once()
	mockConfigStore.On("GetConfig").Return(cfg, nil).Once()
	handler = grpcHandler{mockConfigStore, mockService}
	nftMintResponse, err = handler.MintInvoiceUnpaidNFT(testingconfig.HandlerContext(mockConfigStore), nftReq)
	mockService.AssertExpectations(t)
	assert.Nil(t, err, "mint nft should be successful")
}

func mockmockConfigStore() *configstore.MockService {
	mockConfigStore := &configstore.MockService{}
	mockConfigStore.On("GetAccount", mock.Anything).Return(&configstore.Account{}, nil)
	mockConfigStore.On("GetAllAccounts").Return([]config.Account{&configstore.Account{}}, nil)
	return mockConfigStore
}
