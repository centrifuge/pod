package p2p

import (
	"context"
	"fmt"

	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/centrifuge-protobufs/gen/go/p2p"
	"github.com/centrifuge/go-centrifuge/centerrors"
	"github.com/centrifuge/go-centrifuge/code"
	"github.com/centrifuge/go-centrifuge/coredocument"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/version"
)

// getService looks up the specific registry, derives service from core document
func getServiceAndModel(registry *documents.ServiceRegistry, cd *coredocumentpb.CoreDocument) (documents.Service, documents.Model, error) {
	if cd == nil {
		return nil, nil, fmt.Errorf("nil core document")
	}
	docType, err := coredocument.GetTypeURL(cd)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get type of the document: %v", err)
	}

	srv, err := registry.LocateService(docType)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to locate the service: %v", err)
	}

	model, err := srv.DeriveFromCoreDocument(cd)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to derive model from core document: %v", err)
	}

	return srv, model, nil
}

// handler implements the grpc interface
type handler struct {
	registry *documents.ServiceRegistry
}

// GRPCHandler returns an implementation of P2PServiceServer
func GRPCHandler(registry *documents.ServiceRegistry) p2ppb.P2PServiceServer {
	return handler{registry: registry}
}

// RequestDocumentSignature signs the received document and returns the signature of the signingRoot
// Document signing root will be recalculated and verified
// Existing signatures on the document will be verified
// Document will be stored to the repository for state management
func (srv handler) RequestDocumentSignature(ctx context.Context, sigReq *p2ppb.SignatureRequest) (*p2ppb.SignatureResponse, error) {
	err := handshakeValidator().Validate(sigReq.Header)
	if err != nil {
		return nil, err
	}

	svc, model, err := getServiceAndModel(srv.registry, sigReq.Document)
	if err != nil {
		return nil, centerrors.New(code.DocumentInvalid, err.Error())
	}

	signature, err := svc.RequestDocumentSignature(model)
	if err != nil {
		return nil, centerrors.New(code.Unknown, err.Error())
	}

	return &p2ppb.SignatureResponse{
		CentNodeVersion: version.GetVersion().String(),
		Signature:       signature,
	}, nil
}

// SendAnchoredDocument receives a new anchored document, validates and updates the document in DB
func (srv handler) SendAnchoredDocument(ctx context.Context, docReq *p2ppb.AnchorDocumentRequest) (*p2ppb.AnchorDocumentResponse, error) {
	err := handshakeValidator().Validate(docReq.Header)
	if err != nil {
		return nil, err
	}

	svc, model, err := getServiceAndModel(srv.registry, docReq.Document)
	if err != nil {
		return nil, centerrors.New(code.DocumentInvalid, err.Error())
	}

	err = svc.ReceiveAnchoredDocument(model, docReq.Header)
	if err != nil {
		return nil, centerrors.New(code.Unknown, err.Error())
	}

	return &p2ppb.AnchorDocumentResponse{
		CentNodeVersion: version.GetVersion().String(),
		Accepted:        true,
	}, nil
}
