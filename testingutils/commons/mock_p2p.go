// +build unit integration

package testingcommons

import (
	"context"

	"github.com/centrifuge/centrifuge-protobufs/gen/go/p2p"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
)

// P2PMockClient implements p2ppb.P2PServiceClient
type P2PMockClient struct {
	mock.Mock
}

func (p2p *P2PMockClient) RequestDocumentSignature(ctx context.Context, in *p2ppb.SignatureRequest, opts ...grpc.CallOption) (*p2ppb.SignatureResponse, error) {
	args := p2p.Called(ctx, in, opts)
	resp, _ := args.Get(0).(*p2ppb.SignatureResponse)
	return resp, args.Error(1)
}

func (p2p *P2PMockClient) SendAnchoredDocument(ctx context.Context, in *p2ppb.AnchorDocumentRequest, opts ...grpc.CallOption) (*p2ppb.AnchorDocumentResponse, error) {
	args := p2p.Called(ctx, in, opts)
	resp, _ := args.Get(0).(*p2ppb.AnchorDocumentResponse)
	return resp, args.Error(1)
}
