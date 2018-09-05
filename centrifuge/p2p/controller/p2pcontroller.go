package p2pcontroller

import (
	"context"

	"github.com/CentrifugeInc/centrifuge-protobufs/gen/go/p2p"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/notification"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/p2p/grpcservice"
)

type P2PService struct{}

func (srv *P2PService) Post(ctx context.Context, req *p2ppb.P2PMessage) (rep *p2ppb.P2PReply, err error) {
	// We can define at runtime which notification backend to use pub/sub, queue, webhook ...
	var svc = grpcservice.P2PService{Notifier: &notification.WebhookSender{}}
	return svc.HandleP2PPost(ctx, req)
}

func (srv *P2PService) RequestDocumentSignature(ctx context.Context, sigReq *p2ppb.SignatureRequest) (*p2ppb.SignatureResponse, error) {
	return nil, nil
}

func (srv *P2PService) SendAnchoredDocument(ctx context.Context, docReq *p2ppb.AnchDocumentRequest) (*p2ppb.AnchDocumentResponse, error) {
	return nil, nil
}
