package receiver

import (
	"context"
	"fmt"
	"github.com/centrifuge/go-centrifuge/p2p/common"

	"github.com/centrifuge/go-centrifuge/config/configstore"

	"github.com/golang/protobuf/proto"
	"github.com/libp2p/go-libp2p-peer"
	"github.com/libp2p/go-libp2p-protocol"

	"github.com/centrifuge/go-centrifuge/contextutil"

	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/centrifuge-protobufs/gen/go/p2p"
	"github.com/centrifuge/go-centrifuge/centerrors"
	"github.com/centrifuge/go-centrifuge/code"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/coredocument"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/errors"
	pb "github.com/centrifuge/go-centrifuge/protobufs/gen/go/protocol"
	logging "github.com/ipfs/go-log"
)

var log = logging.Logger("grpc")

// getService looks up the specific registry, derives service from core document
func getServiceAndModel(registry *documents.ServiceRegistry, cd *coredocumentpb.CoreDocument) (documents.Service, documents.Model, error) {
	if cd == nil {
		return nil, nil, errors.New("nil core document")
	}
	docType, err := coredocument.GetTypeURL(cd)
	if err != nil {
		return nil, nil, errors.New("failed to get type of the document: %v", err)
	}

	srv, err := registry.LocateService(docType)
	if err != nil {
		return nil, nil, errors.New("failed to locate the service: %v", err)
	}

	model, err := srv.DeriveFromCoreDocument(cd)
	if err != nil {
		return nil, nil, errors.New("failed to derive model from core document: %v", err)
	}

	return srv, model, nil
}

// Handler implements the grpc interface
type Handler struct {
	registry *documents.ServiceRegistry
	config   config.Configuration
}

// New returns an implementation of P2PServiceServer
func New(config config.Configuration, registry *documents.ServiceRegistry) *Handler {
	return &Handler{registry: registry, config: config}
}

// HandleInterceptor acts as main entry point for all message types, routes the request to the correct handler
func (srv *Handler) HandleInterceptor(ctx context.Context, peer peer.ID, protoc protocol.ID, msg *pb.P2PEnvelope) (*pb.P2PEnvelope, error) {
	envelope, err := p2pcommon.ResolveDataEnvelope(msg)
	if err != nil {
		return nil, err
	}

	err = handshakeValidator(srv.config.GetNetworkID()).Validate(envelope.Header)
	if err != nil {
		return nil, err
	}

	switch p2pcommon.MessageTypeFromString(envelope.Header.Type) {
	case p2pcommon.MessageTypeRequestSignature:
		return srv.HandleRequestDocumentSignature(ctx, peer, protoc, envelope)
	case p2pcommon.MessageTypeSendAnchoredDoc:
		return srv.HandleSendAnchoredDocument(ctx, peer, protoc, envelope)
	default:
		return nil, errors.New("MessageType [%s] not found", envelope.Header.Type)
	}

}

// HandleRequestDocumentSignature handles the RequestDocumentSignature message
func (srv *Handler) HandleRequestDocumentSignature(ctx context.Context, peer peer.ID, protoc protocol.ID, msg *p2ppb.CentrifugeEnvelope) (*pb.P2PEnvelope, error) {
	req := new(p2ppb.SignatureRequest)
	err := proto.Unmarshal(msg.Body, req)
	if err != nil {
		return nil, err
	}

	res, err := srv.RequestDocumentSignature(ctx, req)
	if err != nil {
		return convertToErrorEnvelop(err)
	}

	return p2pcommon.PrepareP2PEnvelope(ctx, srv.config.GetNetworkID(), p2pcommon.MessageTypeRequestSignatureRep, res)
}

// RequestDocumentSignature signs the received document and returns the signature of the signingRoot
// Document signing root will be recalculated and verified
// Existing signatures on the document will be verified
// Document will be stored to the repository for state management
func (srv *Handler) RequestDocumentSignature(ctx context.Context, sigReq *p2ppb.SignatureRequest) (*p2ppb.SignatureResponse, error) {
	// TODO [multi-tenancy] remove following and read the config from the context
	tc, err := configstore.NewTenantConfig("", srv.config)
	if err != nil {
		log.Error(err)
		return nil, centerrors.New(code.Unknown, fmt.Sprintf("failed to get header: %v", err))
	}
	ctxHeader, err := contextutil.NewCentrifugeContext(ctx, tc)
	if err != nil {
		log.Error(err)
		return nil, centerrors.New(code.Unknown, fmt.Sprintf("failed to get header: %v", err))
	}

	svc, model, err := getServiceAndModel(srv.registry, sigReq.Document)
	if err != nil {
		return nil, centerrors.New(code.DocumentInvalid, err.Error())
	}

	signature, err := svc.RequestDocumentSignature(ctxHeader, model)
	if err != nil {
		return nil, centerrors.New(code.Unknown, err.Error())
	}

	return &p2ppb.SignatureResponse{Signature: signature}, nil
}

// HandleSendAnchoredDocument handles the SendAnchoredDocument message
func (srv *Handler) HandleSendAnchoredDocument(ctx context.Context, peer peer.ID, protoc protocol.ID, msg *p2ppb.CentrifugeEnvelope) (*pb.P2PEnvelope, error) {
	m := new(p2ppb.AnchorDocumentRequest)
	err := proto.Unmarshal(msg.Body, m)
	if err != nil {
		return nil, err
	}

	res, err := srv.SendAnchoredDocument(ctx, m, msg.Header.SenderCentrifugeId)
	if err != nil {
		return convertToErrorEnvelop(err)
	}

	return p2pcommon.PrepareP2PEnvelope(ctx, srv.config.GetNetworkID(), p2pcommon.MessageTypeSendAnchoredDocRep, res)
}

// SendAnchoredDocument receives a new anchored document, validates and updates the document in DB
func (srv *Handler) SendAnchoredDocument(ctx context.Context, docReq *p2ppb.AnchorDocumentRequest, senderID []byte) (*p2ppb.AnchorDocumentResponse, error) {
	// TODO [multi-tenancy] remove following and read the config from the context
	tc, err := configstore.NewTenantConfig("", srv.config)
	if err != nil {
		log.Error(err)
		return nil, centerrors.New(code.Unknown, fmt.Sprintf("failed to get header: %v", err))
	}
	ctxHeader, err := contextutil.NewCentrifugeContext(ctx, tc)
	if err != nil {
		log.Error(err)
		return nil, centerrors.New(code.Unknown, fmt.Sprintf("failed to get header: %v", err))
	}

	svc, model, err := getServiceAndModel(srv.registry, docReq.Document)
	if err != nil {
		return nil, centerrors.New(code.DocumentInvalid, err.Error())
	}

	err = svc.ReceiveAnchoredDocument(ctxHeader, model, senderID)
	if err != nil {
		return nil, centerrors.New(code.Unknown, err.Error())
	}

	return &p2ppb.AnchorDocumentResponse{ Accepted: true}, nil
}

func convertToErrorEnvelop(err error) (*pb.P2PEnvelope, error) {
	errPb, ok := err.(proto.Message)
	if !ok {
		return nil, err
	}
	errBytes, err := proto.Marshal(errPb)
	if err != nil {
		return nil, err
	}

	envelope := &p2ppb.CentrifugeEnvelope{
		Header: &p2ppb.CentrifugeHeader{
			Type: p2pcommon.MessageTypeError.String(),
		},
		Body: errBytes,
	}

	marshalledOut, err := proto.Marshal(envelope)
	if err != nil {
		return nil, err
	}

	// an error for the client
	return &pb.P2PEnvelope{Body: marshalledOut}, nil
}
