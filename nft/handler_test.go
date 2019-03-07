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
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockPaymentObligationService struct {
	mock.Mock
}

func (m *mockPaymentObligationService) MintNFT(ctx context.Context, request MintNFTRequest) (*MintNFTResponse, chan bool, error) {
	args := m.Called(ctx, request)
	resp, _ := args.Get(0).(*MintNFTResponse)
	return resp, nil, args.Error(1)
}

func TestNFTMint_success(t *testing.T) {
	nftMintRequest := getTestSetupData()
	mockService := &mockPaymentObligationService{}
	mockConfigStore := mockmockConfigStore()
	docID, _ := hexutil.Decode(nftMintRequest.Identifier)

	tokID := big.NewInt(1)
	nftResponse := &MintNFTResponse{TokenID: tokID.String()}
	req := MintNFTRequest{
		DocumentID:      docID,
		RegistryAddress: common.HexToAddress(nftMintRequest.RegistryAddress),
		DepositAddress:  common.HexToAddress(nftMintRequest.DepositAddress),
		ProofFields:     nftMintRequest.ProofFields,
	}
	mockService.On("MintNFT", mock.Anything, req).Return(nftResponse, nil)
	handler := grpcHandler{mockConfigStore, mockService}
	nftMintResponse, err := handler.MintNFT(testingconfig.HandlerContext(mockConfigStore), nftMintRequest)
	mockService.AssertExpectations(t)
	assert.Nil(t, err, "mint nft should be successful")
	assert.Equal(t, tokID.String(), nftMintResponse.TokenId, "TokenID should have a dummy value")
}

func mockmockConfigStore() *configstore.MockService {
	mockConfigStore := &configstore.MockService{}
	mockConfigStore.On("GetAccount", mock.Anything).Return(&configstore.Account{}, nil)
	mockConfigStore.On("GetAllAccounts").Return([]config.Account{&configstore.Account{}}, nil)
	return mockConfigStore
}

func TestNFTMint_InvalidIdentifier(t *testing.T) {
	nftMintRequest := getTestSetupData()
	nftMintRequest.Identifier = "32321"
	mockConfigStore := mockmockConfigStore()
	mockConfigStore.On("GetAllAccounts").Return(testingconfig.HandlerContext(mockConfigStore))
	handler := grpcHandler{mockConfigStore, &mockPaymentObligationService{}}
	_, err := handler.MintNFT(testingconfig.HandlerContext(mockConfigStore), nftMintRequest)
	assert.Error(t, err, "invalid identifier should throw an error")
}

func TestNFTMint_ServiceError(t *testing.T) {
	nftMintRequest := getTestSetupData()
	mockService := &mockPaymentObligationService{}
	docID, _ := hexutil.Decode(nftMintRequest.Identifier)
	req := MintNFTRequest{
		DocumentID:      docID,
		RegistryAddress: common.HexToAddress(nftMintRequest.RegistryAddress),
		DepositAddress:  common.HexToAddress(nftMintRequest.DepositAddress),
		ProofFields:     nftMintRequest.ProofFields,
	}

	mockService.On("MintNFT", mock.Anything, req).Return(nil, errors.New("service error"))
	mockConfigStore := mockmockConfigStore()
	handler := grpcHandler{mockConfigStore, mockService}
	_, err := handler.MintNFT(testingconfig.HandlerContext(mockConfigStore), nftMintRequest)
	mockService.AssertExpectations(t)
	assert.NotNil(t, err)
}

func TestNFTMint_InvalidAddresses(t *testing.T) {
	nftMintRequest := getTestSetupData()
	nftMintRequest.RegistryAddress = "0x1234"
	mockConfigStore := mockmockConfigStore()
	handler := grpcHandler{mockConfigStore, &mockPaymentObligationService{}}
	_, err := handler.MintNFT(testingconfig.HandlerContext(mockConfigStore), nftMintRequest)
	assert.Error(t, err, "invalid registry address should throw an error")

	nftMintRequest = getTestSetupData()
	nftMintRequest.DepositAddress = "abc"
	handler = grpcHandler{mockConfigStore, &mockPaymentObligationService{}}
	_, err = handler.MintNFT(testingconfig.HandlerContext(mockConfigStore), nftMintRequest)
	assert.Error(t, err, "invalid deposit address should throw an error")
}

func getTestSetupData() *nftpb.NFTMintRequest {
	return &nftpb.NFTMintRequest{
		Identifier:      "0x12121212",
		RegistryAddress: "0xf72855759a39fb75fc7341139f5d7a3974d4da08",
		ProofFields:     []string{"gross_amount", "due_date", "currency"},
		DepositAddress:  "0xf72855759a39fb75fc7341139f5d7a3974d4da08"}
}
