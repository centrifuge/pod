package p2phandler

import (
	"context"
	"fmt"


	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"

	"github.com/centrifuge/centrifuge-protobufs/gen/go/p2p"
	"github.com/centrifuge/go-centrifuge/centrifuge/centerrors"
	"github.com/centrifuge/go-centrifuge/centrifuge/code"
	"github.com/centrifuge/go-centrifuge/centrifuge/coredocument"

	"github.com/centrifuge/go-centrifuge/centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/centrifuge/notification"
	"github.com/centrifuge/go-centrifuge/centrifuge/version"
)

// getService looks up the specific registry, derives service from core document
func getServiceAndModel(cd *coredocumentpb.CoreDocument) (documents.Service, documents.Model, error) {
	if cd == nil {
		return nil, nil, fmt.Errorf("nil core document")
	}
	docType, err := coredocument.GetTypeURL(cd)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get type of the document: %v", err)
	}

	srv, err := documents.GetRegistryInstance().LocateService(docType)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to locate the service: %v", err)
	}

	model, err := srv.DeriveFromCoreDocument(cd)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to derive model from core document: %v", err)
	}

	return srv, model, nil
}

// Handler implements the grpc interface
type Handler struct {
	Notifier notification.Sender
}

// Deprecated
func (srv *Handler) Post(ctx context.Context, req *p2ppb.P2PMessage) (*p2ppb.P2PReply, error) {
	return nil,nil
}

// RequestDocumentSignature signs the received document and returns the signature of the signingRoot
// Document signing root will be recalculated and verified
// Existing signatures on the document will be verified
// Document will be stored to the repository for state management
func (srv *Handler) RequestDocumentSignature(ctx context.Context, sigReq *p2ppb.SignatureRequest) (*p2ppb.SignatureResponse, error) {
	err := handshakeValidator().Validate(sigReq.Header)
	if err != nil {
		return nil, err
	}

	svc, model, err := getServiceAndModel(sigReq.Document)
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
func (srv *Handler) SendAnchoredDocument(ctx context.Context, docReq *p2ppb.AnchDocumentRequest) (*p2ppb.AnchDocumentResponse, error) {
	err := handshakeValidator().Validate(docReq.Header)
	if err != nil {
		return nil, err
	}

	svc, model, err := getServiceAndModel(docReq.Document)
	if err != nil {
		return nil, centerrors.New(code.DocumentInvalid, err.Error())
	}

	err = svc.ReceiveAnchoredDocument(model, docReq.Header)
	if err != nil {
		return nil, centerrors.New(code.Unknown, err.Error())
	}

	return &p2ppb.AnchDocumentResponse{
		CentNodeVersion: version.GetVersion().String(),
		Accepted:        true,
	}, nil
}
