package coredocumentservice

import (
	"context"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/coredocument/repository"
	"github.com/CentrifugeInc/centrifuge-protobufs/gen/go/p2p"
)

type CoreDocumentService struct {}

func (srv *CoreDocumentService) HandleP2PPost(ctx context.Context, req *p2ppb.P2PMessage) (rep *p2ppb.P2PReply, err error) {
	err = coredocumentrepository.GetCoreDocumentRepository().Store(req.Document)
	if err != nil {
		return nil, err
	}

	rep = &p2ppb.P2PReply{Document: req.Document}
	return
}
