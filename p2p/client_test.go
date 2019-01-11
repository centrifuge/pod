// +build unit

package p2p

import (
	"context"
	"testing"

	"github.com/centrifuge/go-centrifuge/p2p/common"

	"github.com/centrifuge/go-centrifuge/testingutils/config"

	"github.com/golang/protobuf/proto"

	"github.com/centrifuge/go-centrifuge/errors"

	"github.com/centrifuge/go-centrifuge/protobufs/gen/go/protocol"
	libp2pPeer "github.com/libp2p/go-libp2p-peer"
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

func (mm *MockMessenger) init(id ...protocol.ID) {
	mm.Called(id)
}

func (mm *MockMessenger) sendMessage(ctx context.Context, p libp2pPeer.ID, pmes *protocolpb.P2PEnvelope, protoc protocol.ID) (*protocolpb.P2PEnvelope, error) {
	args := mm.Called(ctx, p, pmes, protoc)
	resp, _ := args.Get(0).(*protocolpb.P2PEnvelope)
	return resp, args.Error(1)
}

func TestGetSignatureForDocument_fail_connect(t *testing.T) {
	m := &MockMessenger{}
	testClient := &peer{config: cfg, mes: m}
	coreDoc := testingcoredocument.GenerateCoreDocument()
	c, err := cfg.GetConfig()
	assert.NoError(t, err)
	ctx := testingconfig.CreateTenantContext(t, nil, c)

	centrifugeId, err := identity.ToCentID(utils.RandomSlice(identity.CentIDLength))
	assert.Nil(t, err, "centrifugeId not initialized correctly ")

	envelope, err := p2pcommon.PrepareP2PEnvelope(ctx, c.NetworkID, p2pcommon.MessageTypeRequestSignature, &p2ppb.SignatureRequest{Document: coreDoc})
	assert.NoError(t, err, "signature request could not be created")

	m.On("sendMessage", ctx, libp2pPeer.ID("peerID"), envelope, p2pcommon.ProtocolForCID(centrifugeId)).Return(nil, errors.New("some error"))
	resp, err := testClient.getSignatureForDocument(ctx, nil, *coreDoc, "peerID", centrifugeId)
	m.AssertExpectations(t)
	assert.Error(t, err, "must fail")
	assert.Nil(t, resp, "must be nil")
}

func TestGetSignatureForDocument_fail_version_check(t *testing.T) {
	m := &MockMessenger{}
	testClient := &peer{config: cfg, mes: m}
	coreDoc := testingcoredocument.GenerateCoreDocument()
	c, err := cfg.GetConfig()
	assert.NoError(t, err)
	ctx := testingconfig.CreateTenantContext(t, nil, c)

	centrifugeId, err := identity.ToCentID(utils.RandomSlice(identity.CentIDLength))
	assert.Nil(t, err, "centrifugeId not initialized correctly ")

	envelope, err := p2pcommon.PrepareP2PEnvelope(ctx, c.NetworkID, p2pcommon.MessageTypeRequestSignature, &p2ppb.SignatureRequest{Document: coreDoc})
	assert.NoError(t, err, "signature request could not be created")

	m.On("sendMessage", ctx, libp2pPeer.ID("peerID"), envelope, p2pcommon.ProtocolForCID(centrifugeId)).Return(testClient.createSignatureResp("", nil), nil)
	resp, err := testClient.getSignatureForDocument(ctx, nil, *coreDoc, "peerID", centrifugeId)
	m.AssertExpectations(t)
	assert.Error(t, err, "must fail")
	assert.Contains(t, err.Error(), "Incompatible version")
	assert.Nil(t, resp, "must be nil")
}

func TestGetSignatureForDocument_fail_centrifugeId(t *testing.T) {
	m := &MockMessenger{}
	testClient := &peer{config: cfg, mes: m}
	coreDoc := testingcoredocument.GenerateCoreDocument()
	c, err := cfg.GetConfig()
	assert.NoError(t, err)
	ctx := testingconfig.CreateTenantContext(t, nil, c)

	centrifugeId, err := identity.ToCentID(utils.RandomSlice(identity.CentIDLength))
	assert.Nil(t, err, "centrifugeId not initialized correctly ")

	envelope, err := p2pcommon.PrepareP2PEnvelope(ctx, c.NetworkID, p2pcommon.MessageTypeRequestSignature, &p2ppb.SignatureRequest{Document: coreDoc})
	assert.NoError(t, err, "signature request could not be created")

	randomBytes := utils.RandomSlice(identity.CentIDLength)
	signature := &coredocumentpb.Signature{EntityId: randomBytes, PublicKey: utils.RandomSlice(32)}
	m.On("sendMessage", ctx, libp2pPeer.ID("peerID"), envelope, p2pcommon.ProtocolForCID(centrifugeId)).Return(testClient.createSignatureResp(version.GetVersion().String(), signature), nil)

	resp, err := testClient.getSignatureForDocument(ctx, nil, *coreDoc, "peerID", centrifugeId)

	m.AssertExpectations(t)
	assert.Nil(t, resp, "must be nil")
	assert.Error(t, err, "must not be nil")
	assert.Contains(t, err.Error(), "[5]provided bytes doesn't match centID")

}

func (s *peer) createSignatureResp(centNodeVer string, signature *coredocumentpb.Signature) *protocolpb.P2PEnvelope {
	req, err := proto.Marshal(&p2ppb.SignatureResponse{Signature: signature})
	if err != nil {
		return nil
	}

	dataReq := &p2ppb.Envelope{
		Header: &p2ppb.Header{
			NodeVersion: centNodeVer,
			Type:        p2pcommon.MessageTypeRequestSignatureRep.String(),
		},
		Body: req,
	}

	reqB, err := proto.Marshal(dataReq)
	if err != nil {
		return nil
	}

	return &protocolpb.P2PEnvelope{Body: reqB}
}
