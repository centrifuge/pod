package nft

import (
	"context"
	"fmt"

	"github.com/centrifuge/centrifuge-protobufs/documenttypes"
	"github.com/centrifuge/go-centrifuge/centrifuge/centerrors"
	"github.com/centrifuge/go-centrifuge/centrifuge/code"
	"github.com/centrifuge/go-centrifuge/centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/centrifuge/documents/invoice"
	"github.com/centrifuge/go-centrifuge/centrifuge/protobufs/gen/go/nft"
	"github.com/ethereum/go-ethereum/common/hexutil"
	logging "github.com/ipfs/go-log"
)

var apiLog = logging.Logger("document-api")

type grpcHandler struct {
	Service *Service
}

// GRPCHandler returns an implementation of invoice.DocumentServiceServer
func GRPCHandler(service *Service) nftpb.NFTServiceServer {
	if service == nil {
		service = DefaultService()
	}
	return &grpcHandler{Service: service}
}

// MintNFT will be called from the client API to mint a NFT
func (g grpcHandler) MintNFT(context context.Context, request *nftpb.NFTMintRequest) (*nftpb.NFTMintResponse, error) {
	apiLog.Infof("Received request to Mint and NFT", request)
	documentService, err := getDocumentService(request.Identifier)
	if err != nil {
		return &nftpb.NFTMintResponse{}, centerrors.New(code.Unknown, err.Error())
	}

	identifier, err := hexutil.Decode(request.Identifier)
	if err != nil {
		return &nftpb.NFTMintResponse{}, centerrors.New(code.Unknown, err.Error())
	}

	tokenID, err := g.Service.mintNFT(identifier, documentService, request.RegistryAddress, request.DepositAddress, request.ProofFields)
	if err != nil {
		return &nftpb.NFTMintResponse{}, centerrors.New(code.Unknown, err.Error())
	}
	return &nftpb.NFTMintResponse{TokenId: tokenID}, nil
}

func getDocumentService(documentIdentifier string) (invoice.Service, error) {
	// todo concrete service should be returned based on identifier or specific type parameter
	docService, err := documents.GetRegistryInstance().LocateService(documenttypes.InvoiceDataTypeUrl)
	if err != nil {
		return nil, err
	}

	service, ok := docService.(invoice.Service)
	if !ok {
		return nil, fmt.Errorf("couldn't find service for needed document type")

	}
	return service, nil
}
