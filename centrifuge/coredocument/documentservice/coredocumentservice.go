package documentservice

import (
	"context"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/coredocument/grpc"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/coredocument/coredocument_repository"
)

type P2PService struct {}

func (srv *P2PService) Post(ctx context.Context, req *grpc.P2PMessage) (rep *grpc.P2PReply, err error) {
	err = coredocument_repository.GetCoreDocumentRepository().Store(req.Document)
	if err != nil {
		return nil, err
	}

	rep = &grpc.P2PReply{req.Document}
	return
}
