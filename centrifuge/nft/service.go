package nft

import (
	"github.com/centrifuge/go-centrifuge/centrifuge/config"
	"github.com/centrifuge/go-centrifuge/centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/centrifuge/documents/invoice"
	"github.com/centrifuge/go-centrifuge/centrifuge/identity"
	"github.com/ethereum/go-ethereum/common"
)

type IdentityServiceImpl struct{}

func (IdentityServiceImpl) getIdentityAddress() (*common.Address, error) {
	centIDBytes, err := config.Config.GetIdentityId()
	if err != nil {
		return nil, err
	}

	centID, err := identity.ToCentID(centIDBytes)

	if err != nil {
		return nil, err
	}

	ethereumIdentity, err := identity.IDService.LookupIdentityForID(centID)

	if err != nil {
		return nil, err
	}

	address, err := ethereumIdentity.GetIdentityAddress()
	return address, nil

}

type IdentityService interface {
	getIdentityAddress() (*common.Address, error)
}

type Service struct {
	PaymentObligation PaymentObligation
	IdentityService   IdentityService
}

func DefaultService() *Service {
	return &Service{PaymentObligation: getConfiguredPaymentObligation(), IdentityService: IdentityServiceImpl{}}
}

func (s Service) mintNFT(model documents.Model, documentService invoice.Service, registryAddress, depositAddress string, proofFields []string) (string, error) {

	corDoc, err := model.PackCoreDocument()
	if err != nil {
		return "", err
	}
	proofs, err := documentService.CreateProofs(corDoc.DocumentIdentifier, proofFields)

	if err != nil {
		return "", err
	}

	toAddress, err := s.IdentityService.getIdentityAddress()

	if err != nil {
		return "", nil
	}

	requestData, err := NewMintRequestData(toAddress, corDoc.CurrentVersion, proofs.FieldProofs, corDoc.DocumentRoot)

	if err != nil {
		return "", err
	}

	_, err = s.PaymentObligation.Mint(requestData.To, requestData.TokenId, requestData.TokenURI, requestData.AnchorId,
		requestData.MerkleRoot, requestData.Values, requestData.Salts, requestData.Proofs)

	if err != nil {
		return "", err
	}

	return requestData.TokenId.String(), nil

}
