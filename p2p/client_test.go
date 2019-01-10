// +build unit

package p2p

import (
	"context"
	"testing"

	"github.com/centrifuge/go-centrifuge/testingutils/commons"

	"github.com/centrifuge/go-centrifuge/testingutils/config"

	"github.com/centrifuge/go-centrifuge/p2p/receiver"

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

func (mm *MockMessenger) addHandler(mType protocolpb.MessageType, handler func(ctx context.Context, peer libp2pPeer.ID, protoc protocol.ID, msg *protocolpb.P2PEnvelope) (*protocolpb.P2PEnvelope, error)) {
	mm.Called(mType, handler)
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
	testClient := &peer{config: cfg, mes: m, disablePeerStore: true}
	coreDoc := testingcoredocument.GenerateCoreDocument()
	c, err := cfg.GetConfig()
	assert.NoError(t, err)
	ctx := testingconfig.CreateTenantContext(t, c)

	centrifugeId, err := identity.ToCentID(utils.RandomSlice(identity.CentIDLength))
	idService := getIDMocks(centrifugeId)
	assert.Nil(t, err, "centrifugeId not initialized correctly ")

	assert.NoError(t, err)
	sender, err := c.GetIdentityID()
	assert.Nil(t, err, "sender centrifugeId not initialized correctly ")
	r, err := testClient.createSignatureRequest(sender, coreDoc)
	assert.Nil(t, err, "signature request could not be created")

	m.On("sendMessage", ctx, mock.Anything, r, receiver.ProtocolForCID(centrifugeId)).Return(nil, errors.New("some error"))
	resp, err := testClient.getSignatureForDocument(ctx, idService, coreDoc, centrifugeId)
	m.AssertExpectations(t)
	assert.Error(t, err, "must fail")
	assert.Nil(t, resp, "must be nil")
}

func TestGetSignatureForDocument_fail_version_check(t *testing.T) {
	centrifugeId, err := identity.ToCentID(utils.RandomSlice(identity.CentIDLength))
	idService := getIDMocks(centrifugeId)
	m := &MockMessenger{}
	testClient := &peer{config: cfg, mes: m, disablePeerStore: true}
	coreDoc := testingcoredocument.GenerateCoreDocument()
	c, err := cfg.GetConfig()
	assert.NoError(t, err)
	ctx := testingconfig.CreateTenantContext(t, c)
	resp := &p2ppb.SignatureResponse{CentNodeVersion: "1.0.0"}
	assert.Nil(t, err, "centrifugeId not initialized correctly ")

	assert.NoError(t, err)
	sender, err := c.GetIdentityID()
	assert.Nil(t, err, "sender centrifugeId not initialized correctly ")
	r, err := testClient.createSignatureRequest(sender, coreDoc)
	assert.Nil(t, err, "signature request could not be created")

	m.On("sendMessage", ctx, mock.Anything, r, receiver.ProtocolForCID(centrifugeId)).Return(testClient.createSignatureResp("", nil), nil)
	resp, err = testClient.getSignatureForDocument(ctx, idService, coreDoc, centrifugeId)
	m.AssertExpectations(t)
	assert.Error(t, err, "must fail")
	assert.Contains(t, err.Error(), "Incompatible version")
	assert.Nil(t, resp, "must be nil")
}

func TestGetSignatureForDocument_fail_centrifugeId(t *testing.T) {
	m := &MockMessenger{}
	testClient := &peer{config: cfg, mes: m, disablePeerStore: true}
	coreDoc := testingcoredocument.GenerateCoreDocument()
	c, err := cfg.GetConfig()
	assert.NoError(t, err)
	ctx := testingconfig.CreateTenantContext(t, c)

	centrifugeId, err := identity.ToCentID(utils.RandomSlice(identity.CentIDLength))
	idService := getIDMocks(centrifugeId)
	assert.Nil(t, err, "centrifugeId not initialized correctly ")

	assert.NoError(t, err)
	sender, err := c.GetIdentityID()
	assert.Nil(t, err, "sender centrifugeId not initialized correctly ")
	r, err := testClient.createSignatureRequest(sender, coreDoc)
	assert.Nil(t, err, "signature request could not be created")

	randomBytes := utils.RandomSlice(identity.CentIDLength)
	signature := &coredocumentpb.Signature{EntityId: randomBytes, PublicKey: utils.RandomSlice(32)}
	m.On("sendMessage", ctx, mock.Anything, r, receiver.ProtocolForCID(centrifugeId)).Return(testClient.createSignatureResp(version.GetVersion().String(), signature), nil)

	resp, err := testClient.getSignatureForDocument(ctx, idService, coreDoc, centrifugeId)

	m.AssertExpectations(t)
	assert.Nil(t, resp, "must be nil")
	assert.Error(t, err, "must not be nil")
	assert.Contains(t, err.Error(), "[5]provided bytes doesn't match centID")

}

func getIDMocks(centrifugeId identity.CentID) *testingcommons.MockIDService {
	idService := &testingcommons.MockIDService{}
	id := &testingcommons.MockID{}
	id.On("CurrentP2PKey").Return("5dsgvJGnvAfiR3K6HCBc4hcokSfmjj", nil)
	idService.On("LookupIdentityForID", centrifugeId).Return(id, nil)
	return idService
}

func (s *peer) createSignatureResp(centNodeVer string, signature *coredocumentpb.Signature) *protocolpb.P2PEnvelope {
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
