package grpcservice

import (
	"context"
	"fmt"
	"github.com/CentrifugeInc/centrifuge-protobufs/gen/go/p2p"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/config"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/coredocument/repository"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/version"
)

type P2PService struct{}

func (srv *P2PService) HandleP2PPost(ctx context.Context, req *p2ppb.P2PMessage) (rep *p2ppb.P2PReply, err error) {
	// Check call compatibility:
	compatible, err := version.CheckMajorCompatibility(req.CentNodeVersion)
	if err != nil {
		return nil, err
	}
	if !compatible {
		return nil, fmt.Errorf("Incompatible version: this node has version %s", version.CENTRIFUGE_NODE_VERSION)
	}

	if req.NetworkIdentifier != config.Config.GetNetworkID() {
		return nil, fmt.Errorf("Incompatible network identifier: this node is on %d", config.Config.GetNetworkID())
	}

	err = coredocumentrepository.GetCoreDocumentRepository().Store(req.Document)
	if err != nil {
		return nil, err
	}

	rep = &p2ppb.P2PReply{
		CentNodeVersion: version.CENTRIFUGE_NODE_VERSION,
		Document:        req.Document,
	}
	return
}
