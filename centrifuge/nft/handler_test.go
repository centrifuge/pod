// +build unit

package nft

import (
	"context"
	"flag"
	"math/big"
	"os"
	"testing"

	"github.com/centrifuge/centrifuge-protobufs/documenttypes"
	cc "github.com/centrifuge/go-centrifuge/centrifuge/context/testingbootstrap"
	"github.com/centrifuge/go-centrifuge/centrifuge/coredocument/repository"
	"github.com/centrifuge/go-centrifuge/centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/centrifuge/documents/invoice"
	clientinvoicepb "github.com/centrifuge/go-centrifuge/centrifuge/protobufs/gen/go/invoice"
	"github.com/centrifuge/go-centrifuge/centrifuge/protobufs/gen/go/nft"
	"github.com/centrifuge/go-centrifuge/centrifuge/testingutils"
	"github.com/centrifuge/go-centrifuge/centrifuge/testingutils/commons"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/core/types"
)

var invService invoice.Service

func registerInvoiceService() {

	proc := &testingutils.MockCoreDocumentProcessor{}
	proc.On("Anchor", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	proc.On("Send", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	invService = invoice.DefaultService(invoice.GetRepository(), proc)

	documents.GetRegistryInstance().Register(documenttypes.InvoiceDataTypeUrl, invService)

}

func TestMain(m *testing.M) {
	cc.TestIntegrationBootstrap()
	db := cc.GetLevelDBStorage()
	coredocumentrepository.InitLevelDBRepository(db)
	invoice.InitLegacyRepository(db)
	flag.Parse()
	registerInvoiceService()

	result := m.Run()

	cc.TestIntegrationTearDown()
	os.Exit(result)

}

func getTestSetupData() *nftpb.NFTMintRequest {

	nftMintRequest := &nftpb.NFTMintRequest{
		Identifier:      "inv1234",
		RegistryAddress: "0xf72855759a39fb75fc7341139f5d7a3974d4da08",
		ProofFields:     []string{"gross_amount", "due_date", "currency"},
		DepositAddress:  "0xf72855759a39fb75fc7341139f5d7a3974d4da08"}

	return nftMintRequest
}

type MockPaymentObligation struct{}

func (MockPaymentObligation) Mint(opts *bind.TransactOpts, _to common.Address, _tokenId *big.Int, _tokenURI string, _anchorId *big.Int, _merkleRoot [32]byte, _values [3]string, _salts [3][32]byte, _proofs [3][][32]byte) (*types.Transaction, error) {
	return nil, nil
}

func getServiceWithMockedPaymentObligation() *PaymentObligationService {
	return &PaymentObligationService{paymentObligation: MockPaymentObligation{}, identityService: &testingcommons.MockIDService{}}
}

func createInvoiceInDB(t *testing.T) []byte {
	payload := &clientinvoicepb.InvoiceCreatePayload{
		Data: &clientinvoicepb.InvoiceData{
			Sender:      "0x010101010101",
			Recipient:   "0x010203040506",
			Payee:       "0x010203020406",
			GrossAmount: 42,
			ExtraData:   "0x",
			Currency:    "EUR",
		},
	}

	inv, err := invService.DeriveFromCreatePayload(payload)
	_, err = invService.Create(context.Background(), inv)

	corDoc, err := inv.PackCoreDocument()
	assert.Nil(t, err, "model should return a valid core document")

	return corDoc.DocumentIdentifier

}

func TestNFTMint_success(t *testing.T) {
	documentIdentifier := createInvoiceInDB(t)
	nftMintRequest := getTestSetupData()
	nftMintRequest.Identifier = string(documentIdentifier)
	handler := grpcHandler{getServiceWithMockedPaymentObligation()}

	nftMintResponse, err := handler.MintNFT(context.Background(), nftMintRequest)

	assert.Nil(t, err, "mint nft should be successful")
	assert.NotEqual(t, "", nftMintResponse.TokenId, "tokenId should have a dummy value")
}

func TestNFTMint_InvalidIdentifier(t *testing.T) {
	nftMintRequest := getTestSetupData()
	handler := grpcHandler{getServiceWithMockedPaymentObligation()}
	nftMintResponse, err := handler.MintNFT(context.Background(), nftMintRequest)

	assert.Error(t, err, "invalid identifier should throw an error")
	assert.Nil(t, nftMintResponse, "nftMintResponse should be nil")
}

func TestNFTMint_InvalidMintRequest(t *testing.T) {
	handler := grpcHandler{getServiceWithMockedPaymentObligation()}
	nftMintResponse, err := handler.MintNFT(context.Background(), &nftpb.NFTMintRequest{})

	assert.Error(t, err, "empty request should throw an error")
	assert.Nil(t, nftMintResponse, "nftMintResponse should be nil")
}
