package receiver

import (
	"context"
	"fmt"
	"time"

	"github.com/centrifuge/go-substrate-rpc-client/v4/types/codec"

	nftv3 "github.com/centrifuge/go-centrifuge/nft/v3"

	"github.com/centrifuge/go-substrate-rpc-client/v4/types"

	errorspb "github.com/centrifuge/centrifuge-protobufs/gen/go/errors"
	p2ppb "github.com/centrifuge/centrifuge-protobufs/gen/go/p2p"
	pb "github.com/centrifuge/centrifuge-protobufs/gen/go/protocol"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/contextutil"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/errors"
	v2 "github.com/centrifuge/go-centrifuge/identity/v2"
	p2pcommon "github.com/centrifuge/go-centrifuge/p2p/common"
	"github.com/centrifuge/go-centrifuge/utils/timeutils"
	"github.com/golang/protobuf/proto"
	logging "github.com/ipfs/go-log"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/protocol"
)

var log = logging.Logger("p2p-handler")

// Handler implements protocol message handlers
type Handler struct {
	config             config.Service
	handshakeValidator ValidatorGroup
	docSrv             documents.Service
	identityService    v2.Service
	nftService         nftv3.Service
}

// New returns an implementation of P2PServiceServer
func New(
	config config.Service,
	handshakeValidator ValidatorGroup,
	docSrv documents.Service,
	identityService v2.Service,
	nftService nftv3.Service,
) *Handler {
	return &Handler{
		config:             config,
		handshakeValidator: handshakeValidator,
		docSrv:             docSrv,
		identityService:    identityService,
		nftService:         nftService,
	}
}

// HandleInterceptor acts as main entry point for all message types, routes the request to the correct handler
func (srv *Handler) HandleInterceptor(ctx context.Context, peer peer.ID, protoc protocol.ID, msg *pb.P2PEnvelope) (*pb.P2PEnvelope, error) {
	cfg, err := srv.config.GetConfig()
	if err != nil {
		return srv.convertToErrorEnvelop(err)
	}
	defer timeutils.EnsureDelayOperation(time.Now(), cfg.GetP2PResponseDelay())

	if msg == nil {
		return srv.convertToErrorEnvelop(errors.New("nil payload provided"))
	}
	envelope, err := p2pcommon.ResolveDataEnvelope(msg)
	if err != nil {
		return srv.convertToErrorEnvelop(err)
	}

	did, err := p2pcommon.ExtractIdentity(protoc)
	if err != nil {
		return srv.convertToErrorEnvelop(err)
	}

	tc, err := srv.config.GetAccount(did[:])
	if err != nil {
		return srv.convertToErrorEnvelop(err)
	}

	ctx = contextutil.WithAccount(ctx, tc)
	collaborator, err := types.NewAccountID(envelope.Header.SenderId)
	if err != nil {
		return srv.convertToErrorEnvelop(err)
	}
	err = srv.handshakeValidator.Validate(envelope.Header, collaborator, &peer)
	if err != nil {
		return srv.convertToErrorEnvelop(err)
	}

	switch p2pcommon.MessageTypeFromString(envelope.Header.Type) {
	case p2pcommon.MessageTypeRequestSignature:
		return srv.HandleRequestDocumentSignature(ctx, peer, protoc, envelope)
	case p2pcommon.MessageTypeSendAnchoredDoc:
		return srv.HandleSendAnchoredDocument(ctx, peer, protoc, envelope)
	case p2pcommon.MessageTypeGetDoc:
		return srv.HandleGetDocument(ctx, peer, protoc, envelope)
	default:
		return srv.convertToErrorEnvelop(errors.New("MessageType [%s] not found", envelope.Header.Type))
	}
}

// HandleRequestDocumentSignature handles the RequestDocumentSignature message
func (srv *Handler) HandleRequestDocumentSignature(ctx context.Context, peer peer.ID, protoc protocol.ID, msg *p2ppb.Envelope) (*pb.P2PEnvelope, error) {
	req := new(p2ppb.SignatureRequest)
	err := proto.Unmarshal(msg.Body, req)
	if err != nil {
		return srv.convertToErrorEnvelop(err)
	}

	collaborator, err := types.NewAccountID(msg.Header.SenderId)
	if err != nil {
		return srv.convertToErrorEnvelop(err)
	}
	res, err := srv.RequestDocumentSignature(ctx, req, collaborator)
	if err != nil {
		return srv.convertToErrorEnvelop(err)
	}

	nc, err := srv.config.GetConfig()
	if err != nil {
		return srv.convertToErrorEnvelop(err)
	}

	p2pEnv, err := p2pcommon.PrepareP2PEnvelope(ctx, nc.GetNetworkID(), p2pcommon.MessageTypeRequestSignatureRep, res)
	if err != nil {
		return srv.convertToErrorEnvelop(err)
	}

	return p2pEnv, nil
}

// RequestDocumentSignature signs the received document and returns the signature of the signingRoot
// document signing root will be recalculated and verified
// Existing signatures on the document will be verified
// document will be stored to the repository for state management
func (srv *Handler) RequestDocumentSignature(ctx context.Context, sigReq *p2ppb.SignatureRequest, collaborator *types.AccountID) (*p2ppb.SignatureResponse, error) {
	if sigReq == nil || sigReq.Document == nil {
		return nil, errors.New("nil document provided")
	}

	model, err := srv.docSrv.DeriveFromCoreDocument(sigReq.Document)
	if err != nil {
		return nil, errors.New("failed to derive from core doc: %v", err)
	}

	signatures, err := srv.docSrv.RequestDocumentSignature(ctx, model, collaborator)
	if err != nil {
		return nil, err
	}

	return &p2ppb.SignatureResponse{Signatures: signatures}, nil
}

// HandleSendAnchoredDocument handles the SendAnchoredDocument message
func (srv *Handler) HandleSendAnchoredDocument(ctx context.Context, peer peer.ID, protoc protocol.ID, msg *p2ppb.Envelope) (*pb.P2PEnvelope, error) {
	m := new(p2ppb.AnchorDocumentRequest)
	err := proto.Unmarshal(msg.Body, m)
	if err != nil {
		return srv.convertToErrorEnvelop(err)
	}

	collaborator, err := types.NewAccountID(msg.Header.SenderId)
	if err != nil {
		return srv.convertToErrorEnvelop(err)
	}
	res, err := srv.SendAnchoredDocument(ctx, m, collaborator)
	if err != nil {
		return srv.convertToErrorEnvelop(err)
	}

	nc, err := srv.config.GetConfig()
	if err != nil {
		return srv.convertToErrorEnvelop(err)
	}

	p2pEnv, err := p2pcommon.PrepareP2PEnvelope(ctx, nc.GetNetworkID(), p2pcommon.MessageTypeSendAnchoredDocRep, res)
	if err != nil {
		return srv.convertToErrorEnvelop(err)
	}

	return p2pEnv, nil
}

// SendAnchoredDocument receives a new anchored document, validates and updates the document in DB
func (srv *Handler) SendAnchoredDocument(ctx context.Context, docReq *p2ppb.AnchorDocumentRequest, collaborator *types.AccountID) (*p2ppb.AnchorDocumentResponse, error) {
	if docReq == nil || docReq.Document == nil {
		return nil, errors.New("nil document provided")
	}

	model, err := srv.docSrv.DeriveFromCoreDocument(docReq.Document)
	if err != nil {
		return nil, errors.New("failed to derive from core doc: %v", err)
	}

	err = srv.docSrv.ReceiveAnchoredDocument(ctx, model, collaborator)
	if err != nil {
		return nil, err
	}

	return &p2ppb.AnchorDocumentResponse{Accepted: true}, nil
}

// HandleGetDocument handles HandleGetDocument message
func (srv *Handler) HandleGetDocument(ctx context.Context, peer peer.ID, protoc protocol.ID, msg *p2ppb.Envelope) (*pb.P2PEnvelope, error) {
	m := new(p2ppb.GetDocumentRequest)
	err := proto.Unmarshal(msg.Body, m)
	if err != nil {
		return srv.convertToErrorEnvelop(err)
	}

	requesterDID, err := types.NewAccountID(msg.Header.SenderId)
	if err != nil {
		return srv.convertToErrorEnvelop(err)
	}

	res, err := srv.GetDocument(ctx, m, requesterDID)
	if err != nil {
		return srv.convertToErrorEnvelop(err)
	}

	nc, err := srv.config.GetConfig()
	if err != nil {
		return srv.convertToErrorEnvelop(err)
	}

	p2pEnv, err := p2pcommon.PrepareP2PEnvelope(ctx, nc.GetNetworkID(), p2pcommon.MessageTypeGetDocRep, res)
	if err != nil {
		return srv.convertToErrorEnvelop(err)
	}

	return p2pEnv, nil
}

// GetDocument receives document identifier and retrieves the corresponding CoreDocument from the repository
func (srv *Handler) GetDocument(ctx context.Context, docReq *p2ppb.GetDocumentRequest, requester *types.AccountID) (*p2ppb.GetDocumentResponse, error) {
	model, err := srv.docSrv.GetCurrentVersion(ctx, docReq.DocumentIdentifier)
	if err != nil {
		return nil, err
	}

	if err = srv.validateDocumentAccess(ctx, docReq, model, requester); err != nil {
		return nil, err
	}

	cd, err := model.PackCoreDocument()
	if err != nil {
		return nil, err
	}

	return &p2ppb.GetDocumentResponse{Document: cd}, nil
}

// validateDocumentAccess validates the GetDocument request against the AccessType indicated in the request
func (srv *Handler) validateDocumentAccess(ctx context.Context, docReq *p2ppb.GetDocumentRequest, m documents.Document, peer *types.AccountID) error {
	// checks which access type is relevant for the request
	switch docReq.AccessType {
	case p2ppb.AccessType_ACCESS_TYPE_REQUESTER_VERIFICATION:
		if !m.AccountCanRead(peer) {
			return ErrAccessDenied
		}
	case p2ppb.AccessType_ACCESS_TYPE_NFT_OWNER_VERIFICATION:
		return srv.validateNFTAccess(ctx, docReq, m, peer)
	case p2ppb.AccessType_ACCESS_TYPE_ACCESS_TOKEN_VERIFICATION:
		// check the document indicated by the delegating document identifier for the access token
		if docReq.AccessTokenRequest == nil {
			return ErrAccessDenied
		}

		modelWithToken, err := srv.docSrv.GetCurrentVersion(ctx, docReq.AccessTokenRequest.DelegatingDocumentIdentifier)
		if err != nil {
			return err
		}

		err = modelWithToken.ATGranteeCanRead(ctx, srv.docSrv, srv.identityService, docReq.AccessTokenRequest.AccessTokenId, docReq.DocumentIdentifier, peer)
		if err != nil {
			return err
		}
	default:
		return ErrInvalidAccessType
	}
	return nil
}

func (srv *Handler) validateNFTAccess(ctx context.Context, docReq *p2ppb.GetDocumentRequest, m documents.Document, peer *types.AccountID) error {
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

	res, err := srv.nftService.OwnerOf(ctx, &nftv3.OwnerOfRequest{
		CollectionID: collectionID,
		ItemID:       itemID,
	})

	if err != nil {
		return fmt.Errorf("couldn't retrieve NFT owner: %w", err)
	}

	if !res.AccountID.Equal(peer) {
		return ErrAccessDenied
	}

	return nil
}

func (srv *Handler) convertToErrorEnvelop(ierr error) (*pb.P2PEnvelope, error) {
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
