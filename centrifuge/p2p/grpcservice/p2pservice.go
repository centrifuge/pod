package grpcservice

import (
	"context"
	"fmt"
	"github.com/CentrifugeInc/centrifuge-protobufs/gen/go/p2p"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/config"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/coredocument/repository"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/version"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/notification"
	"github.com/golang/protobuf/ptypes"
	"time"
	"github.com/CentrifugeInc/centrifuge-protobufs/gen/go/notification"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/errors"
)

type IncompatibleNetworkError struct {
	clientNetworkId uint32
}

func (e *IncompatibleNetworkError) Error() string {
	return fmt.Sprintf("Incompatible network id: this node is on: %d, client reported: %d", config.Config.GetNetworkID(), e.clientNetworkId)
}

type IncompatibleVersionError struct {
	clientVersion string
}

func (e *IncompatibleVersionError) Error() string {
	return fmt.Sprintf("Incompatible version: this node has version: %s, client reported: %s", version.GetVersion(), e.clientVersion)
}

type P2PService struct {
	Notifier notification.Sender
}

// HandleP2PPost does the basic P2P handshake, stores the document received and sends notification to listener.
// It currently does not do any more processing.
//
// The handshake is currently quite primitive as it only allows the request-server
// to recipient to determine if two versions are compatible. A newer node making a
// request could not decide for itself if the request handshake should succeed or not.
func (srv *P2PService) HandleP2PPost(ctx context.Context, req *p2ppb.P2PMessage) (rep *p2ppb.P2PReply, err error) {
	// Check call compatibility:
	compatible, err := version.CheckMajorCompatibility(req.CentNodeVersion)
	if err != nil {
		return nil, err
	}
	if !compatible {
		return nil, &IncompatibleVersionError{req.CentNodeVersion}
	}

	if req.NetworkIdentifier != config.Config.GetNetworkID() {
		return nil, &IncompatibleNetworkError{req.NetworkIdentifier}
	}

	if req.Document == nil {
		return nil, errors.GenerateNilParameterError(req.Document)
	}

	err = coredocumentrepository.GetCoreDocumentRepository().Store(req.Document)
	if err != nil {
		return nil, err
	}

	ts, err := ptypes.TimestampProto(time.Now().UTC())
	if err != nil {
		return nil, err
	}

	notificationMsg := &notificationpb.NotificationMessage{
		EventType: uint32(notification.RECEIVED_PAYLOAD),
		CentrifugeId: req.SenderCentrifugeId,
		Recorded: ts,
		Document: req.Document,
	}

	// Async until we add queuing
	go srv.Notifier.Send(notificationMsg)
	//

	rep = &p2ppb.P2PReply{
		CentNodeVersion: version.GetVersion().String(),
		Document:        req.Document,
	}
	return
}
