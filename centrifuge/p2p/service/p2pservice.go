package p2pservice

import (
	"github.com/CentrifugeInc/centrifuge-protobufs/gen/go/p2p"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/coredocument/repository"
	"context"
)

type P2PService struct {}

func (srv *P2PService) HandleP2PPost(ctx context.Context, req *p2ppb.P2PMessage) (rep *p2ppb.P2PReply, err error) {
	err = coredocumentrepository.GetCoreDocumentRepository().Store(req.Document)
	if err != nil {
		return nil, err
	}

	rep = &p2ppb.P2PReply{Document: req.Document}
	return
}
