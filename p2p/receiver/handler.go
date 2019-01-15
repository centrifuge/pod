package receiver

import (
	"context"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/documents/genericdoc"
	"github.com/centrifuge/go-centrifuge/p2p/common"

	"github.com/centrifuge/go-centrifuge/contextutil"

	"github.com/golang/protobuf/proto"
	"github.com/libp2p/go-libp2p-peer"
	"github.com/libp2p/go-libp2p-protocol"

	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/centrifuge-protobufs/gen/go/p2p"
	"github.com/centrifuge/go-centrifuge/centerrors"
	"github.com/centrifuge/go-centrifuge/code"
	"github.com/centrifuge/go-centrifuge/coredocument"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/errors"
	pb "github.com/centrifuge/go-centrifuge/protobufs/gen/go/protocol"
)

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

// Handler implements protocol message handlers
type Handler struct {
	registry           *documents.ServiceRegistry
	config             config.Service
	handshakeValidator ValidatorGroup
	genericService 	   genericdoc.Service
}

// New returns an implementation of P2PServiceServer
func New(config config.Service, registry *documents.ServiceRegistry, handshakeValidator ValidatorGroup, genericService genericdoc.Service) *Handler {
	return &Handler{registry: registry, config: config, handshakeValidator: handshakeValidator, genericService: genericService}
}

// HandleInterceptor acts as main entry point for all message types, routes the request to the correct handler
func (srv *Handler) HandleInterceptor(ctx context.Context, peer peer.ID, protoc protocol.ID, msg *pb.P2PEnvelope) (*pb.P2PEnvelope, error) {
	if msg == nil {
		return convertToErrorEnvelop(errors.New("nil payload provided"))
	}
	envelope, err := p2pcommon.ResolveDataEnvelope(msg)
	if err != nil {
		return convertToErrorEnvelop(err)
	}

	cid, err := p2pcommon.ExtractCID(protoc)
	if err != nil {
		return convertToErrorEnvelop(err)
	}

	tc, err := srv.config.GetTenant(cid[:])
	if err != nil {
		return convertToErrorEnvelop(err)
	}

	ctx, err = contextutil.NewCentrifugeContext(ctx, tc)
	if err != nil {
		return convertToErrorEnvelop(err)
	}

	err = srv.handshakeValidator.Validate(envelope)
	if err != nil {
		return convertToErrorEnvelop(err)
	}

	switch p2pcommon.MessageTypeFromString(envelope.Header.Type) {
	case p2pcommon.MessageTypeRequestSignature:
		return srv.HandleRequestDocumentSignature(ctx, peer, protoc, envelope)
	case p2pcommon.MessageTypeSendAnchoredDoc:
		return srv.HandleSendAnchoredDocument(ctx, peer, protoc, envelope)
	//new case p2pcommon.MessageTypeGetAnchoredDoc:
		//return srv.HandleGetAnchoredDocument(ctx, peer, protoc, envelop)
	default:
		return convertToErrorEnvelop(errors.New("MessageType [%s] not found", envelope.Header.Type))
	}

}

// HandleRequestDocumentSignature handles the RequestDocumentSignature message
func (srv *Handler) HandleRequestDocumentSignature(ctx context.Context, peer peer.ID, protoc protocol.ID, msg *p2ppb.Envelope) (*pb.P2PEnvelope, error) {
	req := new(p2ppb.SignatureRequest)
	err := proto.Unmarshal(msg.Body, req)
	if err != nil {
		return convertToErrorEnvelop(err)
	}

	res, err := srv.RequestDocumentSignature(ctx, req)
	if err != nil {
		return convertToErrorEnvelop(err)
	}

	nc, err := srv.config.GetConfig()
	if err != nil {
		return convertToErrorEnvelop(err)
	}

	p2pEnv, err := p2pcommon.PrepareP2PEnvelope(ctx, nc.GetNetworkID(), p2pcommon.MessageTypeRequestSignatureRep, res)
	if err != nil {
		return convertToErrorEnvelop(err)
	}

	return p2pEnv, nil
}

// RequestDocumentSignature signs the received document and returns the signature of the signingRoot
// Document signing root will be recalculated and verified
// Existing signatures on the document will be verified
// Document will be stored to the repository for state management
func (srv *Handler) RequestDocumentSignature(ctx context.Context, sigReq *p2ppb.SignatureRequest) (*p2ppb.SignatureResponse, error) {
	svc, model, err := getServiceAndModel(srv.registry, sigReq.Document)
	if err != nil {
		return nil, centerrors.New(code.DocumentInvalid, err.Error())
	}

	signature, err := svc.RequestDocumentSignature(ctx, model)
	if err != nil {
		return nil, centerrors.New(code.Unknown, err.Error())
	}

	return &p2ppb.SignatureResponse{Signature: signature}, nil
}

// HandleSendAnchoredDocument handles the SendAnchoredDocument message
func (srv *Handler) HandleSendAnchoredDocument(ctx context.Context, peer peer.ID, protoc protocol.ID, msg *p2ppb.Envelope) (*pb.P2PEnvelope, error) {
	m := new(p2ppb.AnchorDocumentRequest)
	err := proto.Unmarshal(msg.Body, m)
	if err != nil {
		return convertToErrorEnvelop(err)
	}

	res, err := srv.SendAnchoredDocument(ctx, m, msg.Header.SenderId)
	if err != nil {
		return convertToErrorEnvelop(err)
	}

	nc, err := srv.config.GetConfig()
	if err != nil {
		return convertToErrorEnvelop(err)
	}

	p2pEnv, err := p2pcommon.PrepareP2PEnvelope(ctx, nc.GetNetworkID(), p2pcommon.MessageTypeSendAnchoredDocRep, res)
	if err != nil {
		return convertToErrorEnvelop(err)
	}

	return p2pEnv, nil
}

// SendAnchoredDocument receives a new anchored document, validates and updates the document in DB
func (srv *Handler) SendAnchoredDocument(ctx context.Context, docReq *p2ppb.AnchorDocumentRequest, senderID []byte) (*p2ppb.AnchorDocumentResponse, error) {
	svc, model, err := getServiceAndModel(srv.registry, docReq.Document)
	if err != nil {
		return nil, centerrors.New(code.DocumentInvalid, err.Error())
	}

	err = svc.ReceiveAnchoredDocument(ctx, model, senderID)
	if err != nil {
		return nil, centerrors.New(code.Unknown, err.Error())
	}

	return &p2ppb.AnchorDocumentResponse{Accepted: true}, nil
}

//METHOD HandleGetAnchoredDocument handles HandleGetAnchoredDocument message
//METHOD GetAnchoredDocument

func convertToErrorEnvelop(err error) (*pb.P2PEnvelope, error) {
	errPb, ok := err.(proto.Message)
	if !ok {
		return nil, err
	}
	errBytes, err := proto.Marshal(errPb)
	if err != nil {
		return nil, err
	}

	envelope := &p2ppb.Envelope{
		Header: &p2ppb.Header{
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
