package nft

import (
	"fmt"
	"github.com/centrifuge/go-centrifuge/centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/centrifuge/documents/invoice"
)

type Service struct {
	PaymentObligation PaymentObligation
}

func DefaultService() *Service {

	return &Service{PaymentObligation:getConfiguredPaymentObligation()}
}


func (s Service) mintNFT(model documents.Model, documentService invoice.Service, registryAddress, depositAddress string, proofFields []string) (string, error) {

	proofs, err := documentService.CreateProofs(proofFields, model)

	if err != nil {
		return "", err
	}

	fmt.Println(proofs)

	corDoc, err := model.PackCoreDocument()


	requestData, err := NewMintRequestData(corDoc.CurrentVersion,proofs,corDoc.DocumentRoot)

	if err != nil {
		return "" ,err
	}

	fmt.Println(requestData)

	_, err = s.PaymentObligation.Mint(requestData.To,requestData.TokenId,requestData.TokenURI,requestData.AnchorId,
		requestData.MerkleRoot,requestData.Values,requestData.Salts,requestData.Proofs)


	if err != nil {
		return "", err
	}

	return requestData.TokenId.String(), nil

}