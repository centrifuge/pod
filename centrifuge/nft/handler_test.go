package nft

import (
	"github.com/centrifuge/go-centrifuge/centrifuge/protobufs/gen/go/nft"
	"testing"
)





func getTestSetupData() *nftpb.NFTMintRequest{

	nftMintRequest := &nftpb.NFTMintRequest{
		Identifier:"inv1234",
		RegistryAddress:"0xf72855759a39fb75fc7341139f5d7a3974d4da08",
		ProofFields:  []string{"gross_amount", "due_date", "currency"},
		DepositAddress:"0xf72855759a39fb75fc7341139f5d7a3974d4da08"}

	return nftMintRequest
}

func TestNFTMint(t *testing.T) {

	/*
	nftMintRequest := getTestSetupData()
	handler := GRPCHandler()

	nftMintResponse, err := handler.MintNFT(context.Background(), nftMintRequest)

	assert.Error(t, err,"mint nft template should not throw an error")
	assert.NotEqual(t,"",nftMintResponse.TokenId,"tokenId should have a dummy value")

	*/
}
