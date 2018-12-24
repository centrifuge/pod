// +build unit

package p2p

import (
	"context"
	"testing"

	"github.com/golang/protobuf/proto"

	"github.com/centrifuge/go-centrifuge/errors"

	"github.com/centrifuge/go-centrifuge/protobufs/gen/go/protocol"
	"github.com/libp2p/go-libp2p-net"
	"github.com/libp2p/go-libp2p-peer"
	"github.com/libp2p/go-libp2p-protocol"

	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/centrifuge-protobufs/gen/go/p2p"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/testingutils/coredocument"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/centrifuge/go-centrifuge/version"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockMessenger struct {
	mock.Mock
}

func (mm *MockMessenger) addHandler(mType protocolpb.MessageType, handler func(ctx context.Context, peer peer.ID, protoc protocol.ID, msg *protocolpb.P2PEnvelope) (*protocolpb.P2PEnvelope, error)) {
	mm.Called(mType, handler)
}

func (mm *MockMessenger) handleNewStream(s net.Stream) {
	mm.Called(s)
}

func (mm *MockMessenger) sendRequest(ctx context.Context, p peer.ID, pmes *protocolpb.P2PEnvelope, protoc protocol.ID) (*protocolpb.P2PEnvelope, error) {
	args := mm.Called(ctx, p, pmes, protoc)
	resp, _ := args.Get(0).(*protocolpb.P2PEnvelope)
	return resp, args.Error(1)
}

func TestGetSignatureForDocument_fail_connect(t *testing.T) {
	m := &MockMessenger{}
	testClient := &p2pServer{config: cfg, mes: m}
	coreDoc := testingcoredocument.GenerateCoreDocument()
	ctx := context.Background()

	centrifugeId, err := identity.ToCentID(utils.RandomSlice(identity.CentIDLength))
	assert.Nil(t, err, "centrifugeId not initialized correctly ")

	sender, err := cfg.GetIdentityID()
	assert.Nil(t, err, "sender centrifugeId not initialized correctly ")
	r, err := testClient.createSignatureRequest(sender, coreDoc)
	assert.Nil(t, err, "signature request could not be created")

	m.On("sendRequest", ctx, peer.ID("peerID"), r, CentrifugeProtocol).Return(nil, errors.New("some error"))
	resp, err := testClient.getSignatureForDocument(ctx, nil, *coreDoc, "peerID", centrifugeId)
	m.AssertExpectations(t)
	assert.Error(t, err, "must fail")
	assert.Nil(t, resp, "must be nil")
}

func TestGetSignatureForDocument_fail_version_check(t *testing.T) {
	m := &MockMessenger{}
	testClient := &p2pServer{config: cfg, mes: m}
	coreDoc := testingcoredocument.GenerateCoreDocument()
	ctx := context.Background()
	resp := &p2ppb.SignatureResponse{CentNodeVersion: "1.0.0"}

	centrifugeId, err := identity.ToCentID(utils.RandomSlice(identity.CentIDLength))
	assert.Nil(t, err, "centrifugeId not initialized correctly ")

	sender, err := cfg.GetIdentityID()
	assert.Nil(t, err, "sender centrifugeId not initialized correctly ")
	r, err := testClient.createSignatureRequest(sender, coreDoc)
	assert.Nil(t, err, "signature request could not be created")

	m.On("sendRequest", ctx, peer.ID("peerID"), r, CentrifugeProtocol).Return(testClient.createSignatureResp("", nil), nil)
	resp, err = testClient.getSignatureForDocument(ctx, nil, *coreDoc, "peerID", centrifugeId)
	m.AssertExpectations(t)
	assert.Error(t, err, "must fail")
	assert.Contains(t, err.Error(), "Incompatible version")
	assert.Nil(t, resp, "must be nil")
}

func TestGetSignatureForDocument_fail_centrifugeId(t *testing.T) {
	m := &MockMessenger{}
	testClient := &p2pServer{config: cfg, mes: m}
	coreDoc := testingcoredocument.GenerateCoreDocument()
	ctx := context.Background()

	centrifugeId, err := identity.ToCentID(utils.RandomSlice(identity.CentIDLength))
	assert.Nil(t, err, "centrifugeId not initialized correctly ")

	sender, err := cfg.GetIdentityID()
	assert.Nil(t, err, "sender centrifugeId not initialized correctly ")
	r, err := testClient.createSignatureRequest(sender, coreDoc)
	assert.Nil(t, err, "signature request could not be created")

	randomBytes := utils.RandomSlice(identity.CentIDLength)
	signature := &coredocumentpb.Signature{EntityId: randomBytes, PublicKey: utils.RandomSlice(32)}
	m.On("sendRequest", ctx, peer.ID("peerID"), r, CentrifugeProtocol).Return(testClient.createSignatureResp(version.GetVersion().String(), signature), nil)

	resp, err := testClient.getSignatureForDocument(ctx, nil, *coreDoc, "peerID", centrifugeId)

	m.AssertExpectations(t)
	assert.Nil(t, resp, "must be nil")
	assert.Error(t, err, "must not be nil")
	assert.Contains(t, err.Error(), "[5]provided bytes doesn't match centID")

}

func (s *p2pServer) createSignatureResp(centNodeVer string, signature *coredocumentpb.Signature) *protocolpb.P2PEnvelope {
	req := &p2ppb.SignatureResponse{
		CentNodeVersion: centNodeVer,
		Signature:       signature,
	}

	reqB, err := proto.Marshal(req)
	if err != nil {
		return nil
	}
	return &protocolpb.P2PEnvelope{Type: protocolpb.MessageType_MESSAGE_TYPE_REQUEST_SIGNATURE_REP, Body: reqB}
}
