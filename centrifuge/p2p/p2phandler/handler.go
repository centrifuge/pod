package p2phandler

import (
	"context"
	"fmt"
	"time"

	"github.com/CentrifugeInc/centrifuge-protobufs/gen/go/notification"
	"github.com/CentrifugeInc/centrifuge-protobufs/gen/go/p2p"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/centerrors"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/code"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/config"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/coredocument/repository"
	centED25519 "github.com/CentrifugeInc/go-centrifuge/centrifuge/keytools/ed25519"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/notification"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/signatures"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/version"
	"github.com/golang/protobuf/ptypes"
)

func incompatibleNetworkError(nodeNetwork uint32) error {
	return centerrors.New(code.NetworkMismatch, fmt.Sprintf("Incompatible network id: node network: %d, client network: %d", config.Config.GetNetworkID(), nodeNetwork))
}

// basicChecks does a network and version check for any incompatibility
func basicChecks(nodeVersion string, networkID uint32) error {
	compatible := version.CheckVersion(nodeVersion)
	if !compatible {
		return version.IncompatibleVersionError(nodeVersion)
	}

	if config.Config.GetNetworkID() != networkID {
		return incompatibleNetworkError(networkID)
	}

	return nil
}

// Handler implements the grpc interface
type Handler struct {
	Notifier notification.Sender
}

// Post does the basic P2P handshake, stores the document received and sends notification to listener.
// It currently does not do any more processing.
//
// The handshake is currently quite primitive as it only allows the request-server
// to recipient to determine if two versions are compatible. A newer node making a
// request could not decide for itself if the request handshake should succeed or not.
func (srv *Handler) Post(ctx context.Context, req *p2ppb.P2PMessage) (*p2ppb.P2PReply, error) {
	err := basicChecks(req.CentNodeVersion, req.NetworkIdentifier)
	if err != nil {
		return nil, err
	}

	if req.Document == nil {
		return nil, centerrors.New(code.DocumentInvalid, centerrors.NilError(req.Document).Error())
	}

	err = coredocumentrepository.GetRepository().Create(req.Document.DocumentIdentifier, req.Document)
	if err != nil {
		return nil, centerrors.New(code.Unknown, err.Error())
	}

	// this should ideally never fail. lets ignore the error
	ts, _ := ptypes.TimestampProto(time.Now().UTC())

	notificationMsg := &notificationpb.NotificationMessage{
		EventType:    uint32(notification.RECEIVED_PAYLOAD),
		CentrifugeId: req.SenderCentrifugeId,
		Recorded:     ts,
		Document:     req.Document,
	}

	// Async until we add queuing
	go srv.Notifier.Send(notificationMsg)

	return &p2ppb.P2PReply{
		CentNodeVersion: version.GetVersion().String(),
		Document:        req.Document,
	}, nil
}

// RequestDocumentSignature signs the received document and returns the signature
//
// How do we verify if we want to sign the document?
// Can we assume that if we are called to sign, we simply sign?
// Or maybe we can check the SenderID against KeyInfo?
func (srv *Handler) RequestDocumentSignature(ctx context.Context, sigReq *p2ppb.SignatureRequest) (*p2ppb.SignatureResponse, error) {
	err := basicChecks(sigReq.Header.CentNodeVersion, sigReq.Header.NetworkIdentifier)
	if err != nil {
		return nil, err
	}

	if sigReq.Document == nil {
		return nil, centerrors.New(code.DocumentInvalid, centerrors.NilError(sigReq.Document).Error())
	}

	idConfig, err := centED25519.GetIDConfig()
	if err != nil {
		return nil, centerrors.New(code.Unknown, fmt.Sprintf("failed to get ID Config: %v", err))
	}

	sig := signatures.Sign(idConfig, sigReq.Document)
	return &p2ppb.SignatureResponse{
		CentNodeVersion: version.GetVersion().String(),
		Signature:       sig,
	}, nil
}

func (srv *Handler) SendAnchoredDocument(ctx context.Context, docReq *p2ppb.AnchDocumentRequest) (*p2ppb.AnchDocumentResponse, error) {
	return nil, nil
}
