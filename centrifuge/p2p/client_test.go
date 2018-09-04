// +build unit

package p2p

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/CentrifugeInc/centrifuge-protobufs/gen/go/coredocument"
	"github.com/CentrifugeInc/centrifuge-protobufs/gen/go/p2p"
	cc "github.com/CentrifugeInc/go-centrifuge/centrifuge/context/testing"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/identity"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/testingutils"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/tools"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/version"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
)

func TestMain(m *testing.M) {
	cc.TestIntegrationBootstrap()
	result := m.Run()
	cc.TestIntegrationTearDown()
	os.Exit(result)
}

// p2pMockClient implements p2ppb.P2PServiceClient
type p2pMockClient struct {
	mock.Mock
}

func (p2p *p2pMockClient) Post(ctx context.Context, in *p2ppb.P2PMessage, opts ...grpc.CallOption) (*p2ppb.P2PReply, error) {
	args := p2p.Called(ctx, in, opts)
	resp, _ := args.Get(0).(*p2ppb.P2PReply)
	return resp, args.Error(1)
}

func (p2p *p2pMockClient) RequestDocumentSignature(ctx context.Context, in *p2ppb.SignatureRequest, opts ...grpc.CallOption) (*p2ppb.SignatureResponse, error) {
	args := p2p.Called(ctx, in, opts)
	resp, _ := args.Get(0).(*p2ppb.SignatureResponse)
	return resp, args.Error(1)
}

func (p2p *p2pMockClient) SendAnchoredDocument(ctx context.Context, in *p2ppb.AnchDocumentRequest, opts ...grpc.CallOption) (*p2ppb.AnchDocumentResponse, error) {
	args := p2p.Called(ctx, in, opts)
	resp, _ := args.Get(0).(*p2ppb.AnchDocumentResponse)
	return resp, args.Error(1)
}

func TestGetSignatureForDocument_fail_connect(t *testing.T) {
	client := &p2pMockClient{}
	coreDoc := testingutils.GenerateCoreDocument()
	ctx := context.Background()
	client.On("RequestDocumentSignature", ctx, mock.Anything, mock.Anything).Return(nil, fmt.Errorf("signature failed")).Once()
	resp, err := getSignatureForDocument(ctx, *coreDoc, client)
	client.AssertExpectations(t)
	assert.Error(t, err, "must fail")
	assert.Nil(t, resp, "must be nil")
}

func TestGetSignatureForDocument_fail_version_check(t *testing.T) {
	client := &p2pMockClient{}
	coreDoc := testingutils.GenerateCoreDocument()
	ctx := context.Background()
	resp := &p2ppb.SignatureResponse{CentNodeVersion: "1.0.0"}
	client.On("RequestDocumentSignature", ctx, mock.Anything, mock.Anything).Return(resp, nil).Once()
	resp, err := getSignatureForDocument(ctx, *coreDoc, client)
	client.AssertExpectations(t)
	assert.Error(t, err, "must fail")
	assert.Contains(t, err.Error(), "Incompatible version")
	assert.Nil(t, resp, "must be nil")
}

func TestGetSignatureForDocument_success(t *testing.T) {
	client := &p2pMockClient{}
	coreDoc := testingutils.GenerateCoreDocument()
	ctx := context.Background()
	sigResp := &p2ppb.SignatureResponse{
		CentNodeVersion: version.GetVersion().String(),
		Signature:       &coredocumentpb.Signature{},
	}

	client.On("RequestDocumentSignature", ctx, mock.Anything, mock.Anything).Return(sigResp, nil).Once()
	resp, err := getSignatureForDocument(ctx, *coreDoc, client)
	client.AssertExpectations(t)
	assert.NotNil(t, resp, "must not be nil")
	assert.Equal(t, resp, sigResp, "must be equal")
	assert.Nil(t, err, "must be nil")
}

func TestGetSignaturesForDocument_fail(t *testing.T) {
	err := GetSignaturesForDocument(nil, nil, nil)
	assert.Error(t, err, "must not be nil")

	coreDoc := testingutils.GenerateCoreDocument()
	centIDs := [][]byte{tools.RandomSlice(identity.CentIdByteLength)}
	srv := &testingutils.MockIDService{}
	srv.On("LookupIdentityForId", centIDs[0]).Return(nil, fmt.Errorf("error"))
	err = GetSignaturesForDocument(coreDoc, srv, centIDs)
	srv.AssertExpectations(t)
	assert.Error(t, err, "must not be nil")
	assert.Contains(t, err.Error(), "failed to get P2P urls")
}
