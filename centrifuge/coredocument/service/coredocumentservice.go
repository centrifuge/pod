package coredocumentservice

import (
	"context"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/coredocument/repository"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/coredocument/grpc"
)

type P2PService struct {}

func (srv *P2PService) Post(ctx context.Context, req *coredocumentgrpc.P2PMessage) (rep *coredocumentgrpc.P2PReply, err error) {
	err = coredocumentrepository.GetCoreDocumentRepository().Store(req.Document)
	if err != nil {
		return nil, err
	}

	rep = &coredocumentgrpc.P2PReply{req.Document}
	return
}
