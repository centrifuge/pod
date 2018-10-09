package nft

import (
	"fmt"
	"github.com/centrifuge/go-centrifuge/centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/centrifuge/documents/invoice"
)

type Service struct {

}


func (Service) mintNFT(model documents.Model, documentService invoice.Service, registryAddress, depositAddress string, proofFields []string) (string, error) {

	proofs, err := documentService.CreateProofs(proofFields, model)

	if err != nil {
		return "", err
	}

	//TODO implement ethereum interaction here
	fmt.Println(proofs)


	return "fakeTokenId", nil

}