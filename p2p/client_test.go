//go:build unit
// +build unit

package p2p

import (
	"context"
	"testing"

	coredocumentpb "github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	p2ppb "github.com/centrifuge/centrifuge-protobufs/gen/go/p2p"
	protocolpb "github.com/centrifuge/centrifuge-protobufs/gen/go/protocol"
	"github.com/centrifuge/go-centrifuge/documents/generic"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/identity"
	p2pcommon "github.com/centrifuge/go-centrifuge/p2p/common"
	testingcommons "github.com/centrifuge/go-centrifuge/testingutils/commons"
	testingconfig "github.com/centrifuge/go-centrifuge/testingutils/config"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/centrifuge/go-centrifuge/version"
	"github.com/golang/protobuf/proto"
	libp2pPeer "github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/protocol"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var did = testingidentity.GenerateRandomDID()

type MockMessenger struct {
	mock.Mock
}

func (mm *MockMessenger) Init(id ...protocol.ID) {
	mm.Called(id)
}

func (mm *MockMessenger) SendMessage(ctx context.Context, p libp2pPeer.ID, pmes *protocolpb.P2PEnvelope, protoc protocol.ID) (*protocolpb.P2PEnvelope, error) {
	args := mm.Called(ctx, p, pmes, protoc)
	resp, _ := args.Get(0).(*protocolpb.P2PEnvelope)
	return resp, args.Error(1)
}

func TestGetSignatureForDocument_fail_connect(t *testing.T) {
	c, err := cfg.GetConfig()
	assert.NoError(t, err)
	c = updateKeys(c)
	ctx := testingconfig.CreateAccountContext(t, c)
	idService := getIDMocks(ctx, did)
	m := &MockMessenger{}
	testClient := &peer{config: cfg, idService: idService, mes: m, disablePeerStore: true}
	model, cd := generic.CreateGenericWithEmbedCD(t, ctx, did, nil)
	_, err = p2pcommon.PrepareP2PEnvelope(ctx, c.GetNetworkID(), p2pcommon.MessageTypeRequestSignature, &p2ppb.SignatureRequest{Document: &cd})
	assert.NoError(t, err, "signature request could not be created")

	m.On("SendMessage", ctx, mock.Anything, mock.Anything, p2pcommon.ProtocolForIdentity(did)).Return(nil, errors.New("some error"))
	resp, err := testClient.getSignatureForDocument(ctx, model, did, did)
	m.AssertExpectations(t)
	assert.Error(t, err, "must fail")
	assert.Nil(t, resp, "must be nil")
}

func TestGetSignatureForDocument_fail_version_check(t *testing.T) {
	c, err := cfg.GetConfig()
	assert.NoError(t, err)
	c = updateKeys(c)
	ctx := testingconfig.CreateAccountContext(t, c)
	idService := getIDMocks(ctx, did)
	m := &MockMessenger{}
	testClient := &peer{config: cfg, idService: idService, mes: m, disablePeerStore: true}
	model, cd := generic.CreateGenericWithEmbedCD(t, ctx, did, nil)
	_, err = p2pcommon.PrepareP2PEnvelope(ctx, c.GetNetworkID(), p2pcommon.MessageTypeRequestSignature, &p2ppb.SignatureRequest{Document: &cd})
	assert.NoError(t, err, "signature request could not be created")

	m.On("SendMessage", ctx, mock.Anything, mock.Anything, p2pcommon.ProtocolForIdentity(did)).Return(testClient.createSignatureResp("", nil), nil)
	resp, err := testClient.getSignatureForDocument(ctx, model, did, did)
	m.AssertExpectations(t)
	assert.Error(t, err, "must fail")
	assert.Contains(t, err.Error(), "Incompatible version")
	assert.Nil(t, resp, "must be nil")
}

func TestGetSignatureForDocument_fail_did(t *testing.T) {
	c, err := cfg.GetConfig()
	assert.NoError(t, err)
	c = updateKeys(c)
	ctx := testingconfig.CreateAccountContext(t, c)
	idService := getIDMocks(ctx, did)
	m := &MockMessenger{}
	testClient := &peer{config: cfg, idService: idService, mes: m, disablePeerStore: true}
	model, cd := generic.CreateGenericWithEmbedCD(t, ctx, did, nil)
	err = model.AddUpdateLog(did)
	assert.NoError(t, err)
	_, err = p2pcommon.PrepareP2PEnvelope(ctx, c.GetNetworkID(), p2pcommon.MessageTypeRequestSignature, &p2ppb.SignatureRequest{Document: &cd})
	assert.NoError(t, err, "signature request could not be created")

	randomBytes := utils.RandomSlice(identity.DIDLength)
	signatures := []*coredocumentpb.Signature{{SignatureId: utils.RandomSlice(52), SignerId: randomBytes, PublicKey: utils.RandomSlice(32)}}
	m.On("SendMessage", ctx, mock.Anything, mock.Anything, p2pcommon.ProtocolForIdentity(did)).Return(testClient.createSignatureResp(version.GetVersion().String(), signatures), nil)

	resp, err := testClient.getSignatureForDocument(ctx, model, did, did)

	m.AssertExpectations(t)
	assert.Nil(t, resp, "must be nil")
	assert.Error(t, err, "must not be nil")
	assert.Contains(t, err.Error(), "signature invalid with err: provided bytes doesn't match centID")

}

func getIDMocks(ctx context.Context, did identity.DID) *testingcommons.MockIdentityService {
	idService := &testingcommons.MockIdentityService{}
	idService.On("CurrentP2PKey", did).Return("QmVf6EN6mkqWejWKW2qPu16XpdG3kJo1T3mhahPB5Se5n1", nil)
	idService.On("Exists", ctx, did).Return(nil)
	return idService
}

func (s *peer) createSignatureResp(centNodeVer string, signatures []*coredocumentpb.Signature) *protocolpb.P2PEnvelope {
	req, err := proto.Marshal(&p2ppb.SignatureResponse{Signatures: signatures})
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
