// +build unit

package nft

import (
	"context"
	"testing"

	"math/big"

	"github.com/centrifuge/go-centrifuge/protobufs/gen/go/nft"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/go-errors/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockPaymentObligationService struct {
	mock.Mock
}

func (m *MockPaymentObligationService) MintNFT(documentID []byte, registryAddress, depositAddress string, proofFields []string) (<-chan *watchTokenMinted, error) {
	args := m.Called(documentID, registryAddress, depositAddress, proofFields)
	return args.Get(0).(chan *watchTokenMinted), args.Error(1)
}

func TestNFTMint_success(t *testing.T) {
	nftMintRequest := getTestSetupData()
	mockService := &MockPaymentObligationService{}
	docID, _ := hexutil.Decode(nftMintRequest.Identifier)

	confirmations := make(chan *watchTokenMinted)
	mockService.
		On("MintNFT", docID, nftMintRequest.RegistryAddress, nftMintRequest.DepositAddress, nftMintRequest.ProofFields).
		Return(confirmations, nil)

	tokID := big.NewInt(1)
	go func() {
		confirmations <- &watchTokenMinted{tokID, nil}
	}()

	handler := grpcHandler{mockService}
	nftMintResponse, err := handler.MintNFT(context.Background(), nftMintRequest)

	mockService.AssertExpectations(t)
	assert.Nil(t, err, "mint nft should be successful")
	assert.Equal(t, tokID.String(), nftMintResponse.TokenId, "TokenID should have a dummy value")
}

func TestNFTMint_InvalidIdentifier(t *testing.T) {
	nftMintRequest := getTestSetupData()
	nftMintRequest.Identifier = "32321"
	handler := grpcHandler{&MockPaymentObligationService{}}
	_, err := handler.MintNFT(context.Background(), nftMintRequest)
	assert.Error(t, err, "invalid identifier should throw an error")
}

func TestNFTMint_ServiceError(t *testing.T) {
	nftMintRequest := getTestSetupData()
	mockService := &MockPaymentObligationService{}
	docID, _ := hexutil.Decode(nftMintRequest.Identifier)
	confirmations := make(chan *watchTokenMinted)
	mockService.
		On("MintNFT", docID, nftMintRequest.RegistryAddress, nftMintRequest.DepositAddress, nftMintRequest.ProofFields).
		Return(confirmations, errors.New("service error"))

	handler := grpcHandler{mockService}
	_, err := handler.MintNFT(context.Background(), nftMintRequest)
	mockService.AssertExpectations(t)
	assert.NotNil(t, err)
}

func TestNFTMint_InvalidAddresses(t *testing.T) {
	nftMintRequest := getTestSetupData()
	nftMintRequest.RegistryAddress = "0x1234"
	handler := grpcHandler{&MockPaymentObligationService{}}
	_, err := handler.MintNFT(context.Background(), nftMintRequest)
	assert.Error(t, err, "invalid registry address should throw an error")

	nftMintRequest = getTestSetupData()
	nftMintRequest.DepositAddress = "abc"
	handler = grpcHandler{&MockPaymentObligationService{}}
	_, err = handler.MintNFT(context.Background(), nftMintRequest)
	assert.Error(t, err, "invalid deposit address should throw an error")
}

func getTestSetupData() *nftpb.NFTMintRequest {
	return &nftpb.NFTMintRequest{
		Identifier:      "0x12121212",
		RegistryAddress: "0xf72855759a39fb75fc7341139f5d7a3974d4da08",
		ProofFields:     []string{"gross_amount", "due_date", "currency"},
		DepositAddress:  "0xf72855759a39fb75fc7341139f5d7a3974d4da08"}
}
