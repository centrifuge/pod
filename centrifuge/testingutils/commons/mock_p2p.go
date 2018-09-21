package testingcommons

import (
	"context"

	"github.com/centrifuge/go-centrifuge/centrifuge/identity"
	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/centrifuge-protobufs/gen/go/p2p"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
)

// P2PMockClient implements p2ppb.P2PServiceClient
type P2PMockClient struct {
	mock.Mock
}

func (p2p *P2PMockClient) Post(ctx context.Context, in *p2ppb.P2PMessage, opts ...grpc.CallOption) (*p2ppb.P2PReply, error) {
	args := p2p.Called(ctx, in, opts)
	resp, _ := args.Get(0).(*p2ppb.P2PReply)
	return resp, args.Error(1)
}

func (p2p *P2PMockClient) RequestDocumentSignature(ctx context.Context, in *p2ppb.SignatureRequest, opts ...grpc.CallOption) (*p2ppb.SignatureResponse, error) {
	args := p2p.Called(ctx, in, opts)
	resp, _ := args.Get(0).(*p2ppb.SignatureResponse)
	return resp, args.Error(1)
}

func (p2p *P2PMockClient) SendAnchoredDocument(ctx context.Context, in *p2ppb.AnchDocumentRequest, opts ...grpc.CallOption) (*p2ppb.AnchDocumentResponse, error) {
	args := p2p.Called(ctx, in, opts)
	resp, _ := args.Get(0).(*p2ppb.AnchDocumentResponse)
	return resp, args.Error(1)
}

type MockP2PWrapperClient struct {
	mock.Mock
	P2PMockClient *P2PMockClient
}

func NewMockP2PWrapperClient() *MockP2PWrapperClient {
	return &MockP2PWrapperClient{
		P2PMockClient: &P2PMockClient{},
	}
}

func (m *MockP2PWrapperClient) OpenClient(target string) (p2ppb.P2PServiceClient, error) {
	m.P2PMockClient = &P2PMockClient{}
	return m.P2PMockClient, nil
}

func (m *MockP2PWrapperClient) GetSignaturesForDocument(ctx context.Context, doc *coredocumentpb.CoreDocument, collaborators []identity.CentID) error {
	m.Called(ctx, doc, collaborators)
	return nil
}
