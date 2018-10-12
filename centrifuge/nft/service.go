package nft

import (
	"fmt"

	"github.com/centrifuge/go-centrifuge/centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/centrifuge/identity"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
)

// TODO remove this when we have a proper dependancy injection mechanism
var service Service

func setService(s Service) {
	service = s
}

func getService() *Service {
	return &service
}

type Config interface {
	GetIdentityId() ([]byte, error)
}

type Service struct {
	paymentObligation PaymentObligation
	identityService   identity.Service
	config            Config
}

func NewService(paymentObligation PaymentObligation, identityService identity.Service, config Config) *Service {
	return &Service{paymentObligation: paymentObligation, identityService: identityService, config: config}
}

func (s *Service) mintNFT(documentID []byte, docType, registryAddress, depositAddress string, proofFields []string) (string, error) {
	documentService, err := getDocumentService(docType)
	if err != nil {
		return "", err
	}

	model, err := documentService.GetLastVersion([]byte(documentID))
	if err != nil {
		return "", err
	}

	corDoc, err := model.PackCoreDocument()
	if err != nil {
		return "", err
	}

	proofs, err := documentService.CreateProofs(documentID, proofFields)
	if err != nil {
		return "", err
	}

	toAddress, err := s.getIdentityAddress()
	if err != nil {
		return "", nil
	}

	requestData, err := NewMintRequest(toAddress, corDoc.CurrentVersion, proofs.FieldProofs, corDoc.DocumentRoot)
	if err != nil {
		return "", err
	}

	_, err = s.paymentObligation.Mint(&bind.TransactOpts{}, requestData.To, requestData.TokenId, requestData.TokenURI, requestData.AnchorId,
		requestData.MerkleRoot, requestData.Values, requestData.Salts, requestData.Proofs)
	if err != nil {
		return "", err
	}

	return requestData.TokenId.String(), nil
}

func (s *Service) getIdentityAddress() (common.Address, error) {
	centIDBytes, err := s.config.GetIdentityId()
	if err != nil {
		return common.Address{}, err
	}

	centID, err := identity.ToCentID(centIDBytes)
	if err != nil {
		return common.Address{}, err
	}

	address, err := s.identityService.GetIdentityAddress(centID)
	if err != nil {
		return common.Address{}, err
	}
	return address, nil
}

func getDocumentService(documentType string) (documents.Service, error) {
	docService, err := documents.GetRegistryInstance().LocateService(documentType)
	if err != nil {
		return nil, err
	}

	service, ok := docService.(documents.Service)
	if !ok {
		return nil, fmt.Errorf("couldn't find service for needed document type")

	}
	return service, nil
}
