package nft

import (
	"fmt"

	"github.com/centrifuge/centrifuge-protobufs/documenttypes"
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

func (s Service) mintNFT(documentID []byte, registryAddress, depositAddress string, proofFields []string) (string, error) {
	// TODO concrete service should be returned based on a document type parameter from the request, for now this is invoice specific
	documentService, err := getDocumentService(documenttypes.InvoiceDataTypeUrl)
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

	toAddress, err := s.IdentityService.getIdentityAddress()
	if err != nil {
		return "", nil
	}

	requestData, err := NewMintRequest(toAddress, corDoc.CurrentVersion, proofs.FieldProofs, corDoc.DocumentRoot)
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

func getDocumentService(documentType string) (invoice.Service, error) {
	docService, err := documents.GetRegistryInstance().LocateService(documentType)
	if err != nil {
		return nil, err
	}

	service, ok := docService.(invoice.Service)
	if !ok {
		return nil, fmt.Errorf("couldn't find service for needed document type")

	}
	return service, nil
}
