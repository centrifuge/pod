package receiver

import (
	"context"
	"fmt"
	"time"

	errorspb "github.com/centrifuge/centrifuge-protobufs/gen/go/errors"
	p2ppb "github.com/centrifuge/centrifuge-protobufs/gen/go/p2p"
	pb "github.com/centrifuge/centrifuge-protobufs/gen/go/protocol"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types/codec"
	"github.com/centrifuge/pod/config"
	"github.com/centrifuge/pod/contextutil"
	"github.com/centrifuge/pod/documents"
	"github.com/centrifuge/pod/errors"
	v2 "github.com/centrifuge/pod/identity/v2"
	nftv3 "github.com/centrifuge/pod/nft/v3"
	p2pcommon "github.com/centrifuge/pod/p2p/common"
	"github.com/centrifuge/pod/utils/timeutils"
	"github.com/golang/protobuf/proto"
	logging "github.com/ipfs/go-log"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/protocol"
)

var log = logging.Logger("p2p-handler")

//go:generate mockery --name Handler --structname HandlerMock --filename handler_mock.go --inpackage

type Handler interface {
	HandleInterceptor(ctx context.Context, peer peer.ID, protoc protocol.ID, msg *pb.P2PEnvelope) (*pb.P2PEnvelope, error)
	HandleRequestDocumentSignature(ctx context.Context, peer peer.ID, protoc protocol.ID, msg *p2ppb.Envelope) (*pb.P2PEnvelope, error)
	RequestDocumentSignature(ctx context.Context, sigReq *p2ppb.SignatureRequest, collaborator *types.AccountID) (*p2ppb.SignatureResponse, error)
	HandleSendAnchoredDocument(ctx context.Context, peer peer.ID, protoc protocol.ID, msg *p2ppb.Envelope) (*pb.P2PEnvelope, error)
	SendAnchoredDocument(ctx context.Context, docReq *p2ppb.AnchorDocumentRequest, collaborator *types.AccountID) (*p2ppb.AnchorDocumentResponse, error)
	HandleGetDocument(ctx context.Context, peer peer.ID, protoc protocol.ID, msg *p2ppb.Envelope) (*pb.P2PEnvelope, error)
	GetDocument(ctx context.Context, docReq *p2ppb.GetDocumentRequest, requester *types.AccountID) (*p2ppb.GetDocumentResponse, error)
}

// handler implements protocol message handlers
type handler struct {
	cfg                config.Configuration
	cfgService         config.Service
	handshakeValidator Validator
	docSrv             documents.Service
	identityService    v2.Service
	nftService         nftv3.Service
}

// NewHandler returns an implementation of P2PServiceServer
func NewHandler(
	cfg config.Configuration,
	cfgService config.Service,
	handshakeValidator ValidatorGroup,
	docSrv documents.Service,
	identityService v2.Service,
	nftService nftv3.Service,
) Handler {
	return &handler{
		cfg:                cfg,
		cfgService:         cfgService,
		handshakeValidator: handshakeValidator,
		docSrv:             docSrv,
		identityService:    identityService,
		nftService:         nftService,
	}
}

// HandleInterceptor acts as main entry point for all message types, routes the request to the correct handler
func (h *handler) HandleInterceptor(ctx context.Context, peerID peer.ID, protocolID protocol.ID, msg *pb.P2PEnvelope) (*pb.P2PEnvelope, error) {
	defer timeutils.EnsureDelayOperation(time.Now(), h.cfg.GetP2PResponseDelay())

	if msg == nil {
		return h.convertToErrorEnvelop(errors.New("nil payload provided"))
	}

	envelope, err := p2pcommon.ResolveDataEnvelope(msg)
	if err != nil {
		return h.convertToErrorEnvelop(err)
	}

	identity, err := p2pcommon.ExtractIdentity(protocolID)
	if err != nil {
		return h.convertToErrorEnvelop(err)
	}

	acc, err := h.cfgService.GetAccount(identity.ToBytes())
	if err != nil {
		return h.convertToErrorEnvelop(err)
	}

	ctx = contextutil.WithAccount(ctx, acc)

	collaborator, err := types.NewAccountID(envelope.GetHeader().GetSenderId())
	if err != nil {
		return h.convertToErrorEnvelop(err)
	}

	err = h.handshakeValidator.Validate(envelope.GetHeader(), collaborator, &peerID)
	if err != nil {
		return h.convertToErrorEnvelop(err)
	}

	switch p2pcommon.MessageTypeFromString(envelope.GetHeader().GetType()) {
	case p2pcommon.MessageTypeRequestSignature:
		return h.HandleRequestDocumentSignature(ctx, peerID, protocolID, envelope)
	case p2pcommon.MessageTypeSendAnchoredDoc:
		return h.HandleSendAnchoredDocument(ctx, peerID, protocolID, envelope)
	case p2pcommon.MessageTypeGetDoc:
		return h.HandleGetDocument(ctx, peerID, protocolID, envelope)
	default:
		return h.convertToErrorEnvelop(errors.New("MessageType [%s] not found", envelope.GetHeader().GetType()))
	}
}

// HandleRequestDocumentSignature handles the RequestDocumentSignature message
func (h *handler) HandleRequestDocumentSignature(ctx context.Context, _ peer.ID, _ protocol.ID, msg *p2ppb.Envelope) (*pb.P2PEnvelope, error) {
	req := new(p2ppb.SignatureRequest)
	err := proto.Unmarshal(msg.GetBody(), req)
	if err != nil {
		return h.convertToErrorEnvelop(err)
	}

	collaborator, err := types.NewAccountID(msg.GetHeader().GetSenderId())
	if err != nil {
		return h.convertToErrorEnvelop(err)
	}
	res, err := h.RequestDocumentSignature(ctx, req, collaborator)
	if err != nil {
		return h.convertToErrorEnvelop(err)
	}

	p2pEnv, err := p2pcommon.PrepareP2PEnvelope(ctx, h.cfg.GetNetworkID(), p2pcommon.MessageTypeRequestSignatureRep, res)
	if err != nil {
		return h.convertToErrorEnvelop(err)
	}

	return p2pEnv, nil
}

// RequestDocumentSignature signs the received document and returns the signature of the signingRoot
// document signing root will be recalculated and verified
// Existing signatures on the document will be verified
// document will be stored to the repository for state management
func (h *handler) RequestDocumentSignature(ctx context.Context, sigReq *p2ppb.SignatureRequest, collaborator *types.AccountID) (*p2ppb.SignatureResponse, error) {
	if sigReq == nil || sigReq.GetDocument() == nil {
		return nil, errors.New("nil document provided")
	}

	model, err := h.docSrv.DeriveFromCoreDocument(sigReq.GetDocument())
	if err != nil {
		return nil, errors.New("failed to derive from core doc: %v", err)
	}

	signatures, err := h.docSrv.RequestDocumentSignature(ctx, model, collaborator)
	if err != nil {
		return nil, err
	}

	return &p2ppb.SignatureResponse{Signatures: signatures}, nil
}

// HandleSendAnchoredDocument handles the SendAnchoredDocument message
func (h *handler) HandleSendAnchoredDocument(ctx context.Context, _ peer.ID, _ protocol.ID, msg *p2ppb.Envelope) (*pb.P2PEnvelope, error) {
	m := new(p2ppb.AnchorDocumentRequest)
	err := proto.Unmarshal(msg.GetBody(), m)
	if err != nil {
		return h.convertToErrorEnvelop(err)
	}

	collaborator, err := types.NewAccountID(msg.GetHeader().GetSenderId())
	if err != nil {
		return h.convertToErrorEnvelop(err)
	}
	res, err := h.SendAnchoredDocument(ctx, m, collaborator)
	if err != nil {
		return h.convertToErrorEnvelop(err)
	}

	p2pEnv, err := p2pcommon.PrepareP2PEnvelope(ctx, h.cfg.GetNetworkID(), p2pcommon.MessageTypeSendAnchoredDocRep, res)
	if err != nil {
		return h.convertToErrorEnvelop(err)
	}

	return p2pEnv, nil
}

// SendAnchoredDocument receives a new anchored document, validates and updates the document in DB
func (h *handler) SendAnchoredDocument(ctx context.Context, docReq *p2ppb.AnchorDocumentRequest, collaborator *types.AccountID) (*p2ppb.AnchorDocumentResponse, error) {
	if docReq == nil || docReq.GetDocument() == nil {
		return nil, errors.New("nil document provided")
	}

	model, err := h.docSrv.DeriveFromCoreDocument(docReq.GetDocument())
	if err != nil {
		return nil, errors.New("failed to derive from core doc: %v", err)
	}

	err = h.docSrv.ReceiveAnchoredDocument(ctx, model, collaborator)
	if err != nil {
		return nil, err
	}

	return &p2ppb.AnchorDocumentResponse{Accepted: true}, nil
}

// HandleGetDocument handles HandleGetDocument message
func (h *handler) HandleGetDocument(ctx context.Context, _ peer.ID, _ protocol.ID, msg *p2ppb.Envelope) (*pb.P2PEnvelope, error) {
	m := new(p2ppb.GetDocumentRequest)
	err := proto.Unmarshal(msg.GetBody(), m)
	if err != nil {
		return h.convertToErrorEnvelop(err)
	}

	requester, err := types.NewAccountID(msg.GetHeader().GetSenderId())
	if err != nil {
		return h.convertToErrorEnvelop(err)
	}

	res, err := h.GetDocument(ctx, m, requester)
	if err != nil {
		return h.convertToErrorEnvelop(err)
	}

	p2pEnv, err := p2pcommon.PrepareP2PEnvelope(ctx, h.cfg.GetNetworkID(), p2pcommon.MessageTypeGetDocRep, res)
	if err != nil {
		return h.convertToErrorEnvelop(err)
	}

	return p2pEnv, nil
}

// GetDocument receives document identifier and retrieves the corresponding CoreDocument from the repository
func (h *handler) GetDocument(ctx context.Context, docReq *p2ppb.GetDocumentRequest, requester *types.AccountID) (*p2ppb.GetDocumentResponse, error) {
	model, err := h.docSrv.GetCurrentVersion(ctx, docReq.GetDocumentIdentifier())
	if err != nil {
		return nil, err
	}

	if err = h.validateDocumentAccess(ctx, docReq, model, requester); err != nil {
		return nil, err
	}

	cd, err := model.PackCoreDocument()
	if err != nil {
		return nil, err
	}

	return &p2ppb.GetDocumentResponse{Document: cd}, nil
}

// validateDocumentAccess validates the GetDocument request against the AccessType indicated in the request
func (h *handler) validateDocumentAccess(
	ctx context.Context,
	req *p2ppb.GetDocumentRequest,
	document documents.Document,
	requester *types.AccountID,
) error {
	// checks which access type is relevant for the request
	switch req.AccessType {
	case p2ppb.AccessType_ACCESS_TYPE_REQUESTER_VERIFICATION:
		if !document.AccountCanRead(requester) {
			return ErrAccessDenied
		}
	case p2ppb.AccessType_ACCESS_TYPE_NFT_OWNER_VERIFICATION:
		return h.validateNFTAccess(req, document, requester)
	case p2ppb.AccessType_ACCESS_TYPE_ACCESS_TOKEN_VERIFICATION:
		// check the document indicated by the delegating document identifier for the access token
		if req.GetAccessTokenRequest() == nil {
			return ErrAccessDenied
		}

		modelWithToken, err := h.docSrv.GetCurrentVersion(ctx, req.GetAccessTokenRequest().GetDelegatingDocumentIdentifier())
		if err != nil {
			return err
		}

		err = modelWithToken.ATGranteeCanRead(
			ctx,
			h.docSrv,
			h.identityService,
			req.GetAccessTokenRequest().GetAccessTokenId(),
			req.GetDocumentIdentifier(),
			requester,
		)

		if err != nil {
			return err
		}
	default:
		return ErrInvalidAccessType
	}
	return nil
}

func (h *handler) validateNFTAccess(docReq *p2ppb.GetDocumentRequest, m documents.Document, peer *types.AccountID) error {
	if !m.AccountCanRead(peer) {
		return ErrAccessDenied
	}

	if !m.NFTCanRead(docReq.GetNftCollectionId(), docReq.GetNftItemId()) {
		return ErrAccessDenied
	}

	var collectionID types.U64

	if err := codec.Decode(docReq.GetNftCollectionId(), &collectionID); err != nil {
		return fmt.Errorf("couldn't decode NFT collection ID: %w", err)
	}

	var itemID types.U128

	if err := codec.Decode(docReq.GetNftItemId(), &itemID); err != nil {
		return fmt.Errorf("couldn't decode NFT item ID: %w", err)
	}

	owner, err := h.nftService.GetNFTOwner(collectionID, itemID)

	if err != nil {
		return fmt.Errorf("couldn't retrieve NFT owner: %w", err)
	}

	if !owner.Equal(peer) {
		return ErrAccessDenied
	}

	return nil
}

func (h *handler) convertToErrorEnvelop(ierr error) (*pb.P2PEnvelope, error) {
	// Log on server side
	log.Error(ierr)

	ierr = errors.Mask(ierr)
	errPb := &errorspb.Error{Message: ierr.Error()}
	errBytes, errx := proto.Marshal(errPb)
	if errx != nil {
		return nil, errx
	}

	envelope := &p2ppb.Envelope{
		Header: &p2ppb.Header{
			Type: p2pcommon.MessageTypeError.String(),
		},
		Body: errBytes,
	}

	marshalledOut, errx := proto.Marshal(envelope)
	if errx != nil {
		return nil, errx
	}

	// an error for the client
	return &pb.P2PEnvelope{Body: marshalledOut}, nil
}
