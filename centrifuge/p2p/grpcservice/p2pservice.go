package grpcservice

import (
	"context"
	"fmt"
	"time"

	"github.com/CentrifugeInc/centrifuge-protobufs/gen/go/coredocument"
	"github.com/CentrifugeInc/centrifuge-protobufs/gen/go/notification"
	"github.com/CentrifugeInc/centrifuge-protobufs/gen/go/p2p"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/code"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/config"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/coredocument/repository"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/errors"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/identity"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/notification"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/p2p"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/signatures"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/version"
	"github.com/golang/protobuf/ptypes"
)

func incompatibleVersionError(nodeVersion string) error {
	return errors.New(code.VersionMismatch, fmt.Sprintf("Incompatible version: node version: %s, client version: %s", version.GetVersion(), nodeVersion))
}

func incompatibleNetworkError(nodeNetwork uint32) error {
	return errors.New(code.NetworkMismatch, fmt.Sprintf("Incompatible network id: node network: %d, client network: %d", config.Config.GetNetworkID(), nodeNetwork))
}

type P2PService struct {
	Notifier notification.Sender
}

// checkVersion checks if the peer node version matches with the current node
func checkVersion(peerVersion string) bool {
	compatible, err := version.CheckMajorCompatibility(peerVersion)
	if err != nil {
		return false
	}

	return compatible
}

// HandleP2PPost does the basic P2P handshake, stores the document received and sends notification to listener.
// It currently does not do any more processing.
//
// The handshake is currently quite primitive as it only allows the request-server
// to recipient to determine if two versions are compatible. A newer node making a
// request could not decide for itself if the request handshake should succeed or not.
func (srv *P2PService) HandleP2PPost(ctx context.Context, req *p2ppb.P2PMessage) (rep *p2ppb.P2PReply, err error) {
	// Check call compatibility:
	compatible := checkVersion(req.CentNodeVersion)
	if !compatible {
		return nil, incompatibleVersionError(req.CentNodeVersion)
	}

	if req.NetworkIdentifier != config.Config.GetNetworkID() {
		return nil, incompatibleNetworkError(req.NetworkIdentifier)
	}

	if req.Document == nil {
		return nil, errors.New(code.DocumentInvalid, errors.NilError(req.Document).Error())
	}

	err = coredocumentrepository.GetRepository().Create(req.Document.DocumentIdentifier, req.Document)
	if err != nil {
		return nil, errors.New(code.Unknown, err.Error())
	}

	ts, err := ptypes.TimestampProto(time.Now().UTC())
	if err != nil {
		return nil, errors.New(code.Unknown, err.Error())
	}

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

// getSignatureForDocument requests the target node to sign the document
func getSignatureForDocument(ctx context.Context, req *p2ppb.SignatureRequest, client p2ppb.P2PServiceClient) (*p2ppb.SignatureResponse, error) {
	resp, err := client.RequestDocumentSignature(ctx, req)
	if err != nil {
		return nil, errors.Wrap(err, "request for document signature failed")
	}

	return resp, nil
}

func newSignatureReq(doc *coredocumentpb.CoreDocument) *p2ppb.SignatureRequest {
	header := p2ppb.CentrifugeHeader{
		NetworkIdentifier:  config.Config.GetNetworkID(),
		CentNodeVersion:    version.GetVersion().String(),
		SenderCentrifugeId: config.Config.GetIdentityId(),
	}

	var coreDoc *coredocumentpb.CoreDocument
	*coreDoc = *doc
	return &p2ppb.SignatureRequest{
		Header:   &header,
		Document: coreDoc,
	}
}

// GetSignaturesForDocument requests peer nodes for the signature and verifies them
func GetSignaturesForDocument(doc *coredocumentpb.CoreDocument, idService identity.Service, centIDs [][]byte) error {
	if doc == nil {
		return errors.NilError(doc)
	}

	req := newSignatureReq(doc)
	targets, err := identity.GetClientsP2PURLs(idService, centIDs)
	if err != nil {
		return errors.Wrap(err, "failed to get P2P urls")
	}

	for _, target := range targets {
		client, err := p2p.OpenClient(target)
		if err != nil {
			return errors.Wrap(err, "failed to connect to target")
		}

		// for now going with context.background, once we have a timeout for request
		// we can use context.Timeout for that
		resp, err := getSignatureForDocument(context.Background(), req, client)
		if err != nil {
			return errors.Wrap(err, "failed to get signature")
		}

		compatible := checkVersion(resp.CentNodeVersion)
		if !compatible {
			return incompatibleVersionError(resp.CentNodeVersion)
		}

		ss := signatures.GetSigningService()
		valid, err := ss.ValidateSignature(resp.Signature, doc.SigningRoot)
		if err != nil {
			return errors.Wrap(err, "failed to validate signature")
		}

		if !valid {
			return errors.New(code.AuthenticationFailed, "signature invalid")
		}

		doc.Signatures = append(doc.Signatures, resp.Signature)
	}

	return nil
}
