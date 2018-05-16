package coredocumentcontroller

import (
	"github.com/CentrifugeInc/centrifuge-protobufs/gen/go/p2p"
	"context"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/coredocument/service"
)

type P2PService struct {}

func (srv *P2PService) Post(ctx context.Context, req *p2ppb.P2PMessage) (rep *p2ppb.P2PReply, err error) {
	var svc = coredocumentservice.CoreDocumentService{}
	return svc.HandleP2PPost(ctx, req)
}