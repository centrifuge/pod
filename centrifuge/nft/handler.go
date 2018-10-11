package nft

import (
	"context"
	"fmt"
	"github.com/centrifuge/centrifuge-protobufs/documenttypes"
	"github.com/centrifuge/go-centrifuge/centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/centrifuge/documents/invoice"
	nftpb"github.com/centrifuge/go-centrifuge/centrifuge/protobufs/gen/go/nft"


)

type grpcHandler struct {
	Service *Service
}


// GRPCHandler returns an implementation of invoice.DocumentServiceServer
func GRPCHandler(service *Service) (nftpb.NFTServiceServer) {

	if service == nil {
		service = DefaultService()
	}

	return &grpcHandler{Service:service}
}


func getDocumentService(documentIdentifier string) (invoice.Service, error){

	// todo concrete service should be returned based on identifier or specific type parameter
	modelDeriver, err := documents.GetRegistryInstance().LocateService(documenttypes.InvoiceDataTypeUrl)

	if err != nil {
		return nil, err
	}

	if service, ok := modelDeriver.(invoice.Service); ok {

		return service, nil
	}

	return nil, fmt.Errorf("couldn't return service for needed document type")


}

func (g grpcHandler)MintNFT(context context.Context,request *nftpb.NFTMintRequest) (*nftpb.NFTMintResponse, error) {

	documentService, err := getDocumentService(request.Identifier)

	if err != nil {
		return nil, err
	}

	model, err := documentService.GetLastVersion([]byte(request.Identifier))

	if err != nil {
		return nil, err
	}


	tokenID, err := g.Service.mintNFT(model,documentService,request.RegistryAddress,request.DepositAddress,request.ProofFields)

	return &nftpb.NFTMintResponse{TokenId:tokenID}, err

}


