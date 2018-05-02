package documentservice

import (
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/coredocument/repository"
	"context"
	cc "github.com/CentrifugeInc/go-centrifuge/centrifuge/context"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/coredocument/grpc"
)

type P2PService struct {}

func (srv *P2PService) Post(ctx context.Context, req *grpc.P2PMessage) (rep *grpc.P2PReply, err error) {
	err = repository.NewLevelDBCoreDocumentRepository(cc.LevelDB).Store(req.Document)
	if err != nil {
		return nil, err
	}

	rep = &grpc.P2PReply{req.Document}
	return
}
