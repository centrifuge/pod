// +build unit

package receiver

import (
	"context"
	"crypto/rand"
	"os"
	"testing"

	"github.com/centrifuge/centrifuge-protobufs/gen/go/p2p"
	"github.com/centrifuge/go-centrifuge/anchors"
	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/bootstrap/bootstrappers/testlogging"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/config/configstore"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/ethereum"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/p2p/common"
	"github.com/centrifuge/go-centrifuge/protobufs/gen/go/protocol"
	"github.com/centrifuge/go-centrifuge/queue"
	"github.com/centrifuge/go-centrifuge/storage/leveldb"
	testingcommons "github.com/centrifuge/go-centrifuge/testingutils/commons"
	"github.com/centrifuge/go-centrifuge/testingutils/config"
	"github.com/centrifuge/go-centrifuge/testingutils/documents"
	"github.com/centrifuge/go-centrifuge/transactions/txv1"
	"github.com/centrifuge/go-centrifuge/version"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/golang/protobuf/proto"
	"github.com/libp2p/go-libp2p-crypto"
	libp2pPeer "github.com/libp2p/go-libp2p-peer"
	"github.com/libp2p/go-libp2p-protocol"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var (
	handler       *Handler
	registry      *documents.ServiceRegistry
	cfg           config.Configuration
	mockIDService *testingcommons.MockIdentityService
	defaultPID    libp2pPeer.ID
)

func TestMain(m *testing.M) {
	ctx := make(map[string]interface{})
	ethClient := &testingcommons.MockEthClient{}
	ethClient.On("GetEthClient").Return(nil)
	ctx[ethereum.BootstrappedEthereumClient] = ethClient
	ibootstappers := []bootstrap.TestBootstrapper{
		&testlogging.TestLoggingBootstrapper{},
		&config.Bootstrapper{},
		&leveldb.Bootstrapper{},
		&configstore.Bootstrapper{},
		&queue.Bootstrapper{},
		txv1.Bootstrapper{},
		&anchors.Bootstrapper{},
		documents.Bootstrapper{},
	}
	mockIDService = &testingcommons.MockIdentityService{}
	ctx[identity.BootstrappedDIDService] = mockIDService
	ctx[identity.BootstrappedDIDFactory] = &testingcommons.MockIdentityFactory{}
	bootstrap.RunTestBootstrappers(ibootstappers, ctx)
	cfg = ctx[bootstrap.BootstrappedConfig].(config.Configuration)
	cfgService := ctx[config.BootstrappedConfigStorage].(config.Service)
	registry = ctx[documents.BootstrappedRegistry].(*documents.ServiceRegistry)
	docSrv := documents.DefaultService(nil, nil, registry, mockIDService)
	_, pub, _ := crypto.GenerateEd25519Key(rand.Reader)
	defaultPID, _ = libp2pPeer.IDFromPublicKey(pub)
	mockIDService.On("ValidateKey", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	handler = New(cfgService, HandshakeValidator(cfg.GetNetworkID(), mockIDService), docSrv, new(testingdocuments.MockRegistry), mockIDService)
	result := m.Run()
	bootstrap.RunTestTeardown(ibootstappers)
	os.Exit(result)
}

func TestHandler_RequestDocumentSignature_nilDocument(t *testing.T) {
	req := &p2ppb.SignatureRequest{}
	resp, err := handler.RequestDocumentSignature(context.Background(), req)
	assert.Error(t, err, "must return error")
	assert.Nil(t, resp, "must be nil")
}

func TestHandler_HandleInterceptor_nilPayload(t *testing.T) {
	resp, err := handler.HandleInterceptor(context.Background(), libp2pPeer.ID("SomePeer"), protocol.ID("protocolX"), nil)
	assert.Error(t, err, "must return error")
	assert.Contains(t, err.Error(), "nil payload provided")
	assert.Nil(t, resp, "must be nil")
}

func TestHandler_HandleInterceptor_HeaderEmpty(t *testing.T) {
	resp, err := handler.HandleInterceptor(context.Background(), libp2pPeer.ID("SomePeer"), protocol.ID("protocolX"), &protocolpb.P2PEnvelope{})
	assert.Error(t, err, "must return error")
	assert.Contains(t, err.Error(), "Header field is empty")
	assert.Nil(t, resp, "must be nil")
}

func TestHandler_HandleInterceptor_CentIDNotHex(t *testing.T) {
	ctx := testingconfig.CreateAccountContext(t, cfg)
	p2pEnv, err := p2pcommon.PrepareP2PEnvelope(ctx, cfg.GetNetworkID(), p2pcommon.MessageTypeRequestSignature, &protocolpb.P2PEnvelope{})
	assert.NoError(t, err)
	resp, err := handler.HandleInterceptor(context.Background(), libp2pPeer.ID("SomePeer"), protocol.ID("protocolX"), p2pEnv)
	assert.Error(t, err, "must return error")
	assert.Equal(t, err, identity.ErrMalformedAddress)
	assert.Nil(t, resp, "must be nil")
}

func TestHandler_HandleInterceptor_TenantNotFound(t *testing.T) {
	ctx := testingconfig.CreateAccountContext(t, cfg)
	p2pEnv, err := p2pcommon.PrepareP2PEnvelope(ctx, cfg.GetNetworkID(), p2pcommon.MessageTypeRequestSignature, &protocolpb.P2PEnvelope{})
	assert.NoError(t, err)
	resp, err := handler.HandleInterceptor(context.Background(), libp2pPeer.ID("SomePeer"), protocol.ID("0x89b0a86583c4442acfd71b463e0d3c55ae1412a5"), p2pEnv)
	assert.Error(t, err, "must return error")
	assert.Contains(t, err.Error(), "model not found in db")
	assert.Nil(t, resp, "must be nil")
}

func TestHandler_HandleInterceptor_HandshakeValidationFail(t *testing.T) {
	ctx := testingconfig.CreateAccountContext(t, cfg)
	p2pEnv, err := p2pcommon.PrepareP2PEnvelope(ctx, cfg.GetNetworkID(), p2pcommon.MessageTypeRequestSignature, &protocolpb.P2PEnvelope{})
	assert.NoError(t, err)

	// Manipulate version in Header
	dataEnv, _ := p2pcommon.ResolveDataEnvelope(p2pEnv)
	dataEnv.Header.NodeVersion = "incompatible"
	marshalledRequest, err := proto.Marshal(dataEnv)
	assert.NoError(t, err)
	p2pEnv = &protocolpb.P2PEnvelope{Body: marshalledRequest}

	id, _ := cfg.GetIdentityID()
	resp, err := handler.HandleInterceptor(context.Background(), libp2pPeer.ID("SomePeer"), protocol.ID(hexutil.Encode(id)), p2pEnv)
	assert.Error(t, err, "must return error")
	assert.Contains(t, err.Error(), "Incompatible version")
	assert.Nil(t, resp, "must be nil")

	// Manipulate network in Header
	p2pEnv, err = p2pcommon.PrepareP2PEnvelope(ctx, uint32(999), p2pcommon.MessageTypeRequestSignature, &protocolpb.P2PEnvelope{})
	assert.NoError(t, err)

	resp, err = handler.HandleInterceptor(context.Background(), libp2pPeer.ID("SomePeer"), protocol.ID(hexutil.Encode(id)), p2pEnv)
	assert.Error(t, err, "must return error")
	assert.Contains(t, err.Error(), "Incompatible network id")
	assert.Nil(t, resp, "must be nil")
}

func TestHandler_HandleInterceptor_UnsupportedMessageType(t *testing.T) {
	ctx := testingconfig.CreateAccountContext(t, cfg)
	p2pEnv, err := p2pcommon.PrepareP2PEnvelope(ctx, cfg.GetNetworkID(), p2pcommon.MessageTypeRequestSignature, &protocolpb.P2PEnvelope{})
	assert.NoError(t, err)

	// Manipulate message type in Header + Signature
	dataEnv, _ := p2pcommon.ResolveDataEnvelope(p2pEnv)
	dataEnv.Header.Type = "UnsupportedType"
	marshalledRequest, err := proto.Marshal(dataEnv)
	assert.NoError(t, err)
	p2pEnv = &protocolpb.P2PEnvelope{Body: marshalledRequest}

	id, _ := cfg.GetIdentityID()
	resp, err := handler.HandleInterceptor(context.Background(), defaultPID, protocol.ID(hexutil.Encode(id)), p2pEnv)
	assert.Error(t, err, "must return error")
	assert.Contains(t, err.Error(), "MessageType [UnsupportedType] not found")
	assert.Nil(t, resp, "must be nil")
}

func TestHandler_HandleInterceptor_NilDocument(t *testing.T) {
	ctx := testingconfig.CreateAccountContext(t, cfg)
	p2pEnv, err := p2pcommon.PrepareP2PEnvelope(ctx, cfg.GetNetworkID(), p2pcommon.MessageTypeRequestSignature, &protocolpb.P2PEnvelope{})
	assert.NoError(t, err)

	id, _ := cfg.GetIdentityID()
	resp, err := handler.HandleInterceptor(context.Background(), defaultPID, protocol.ID(hexutil.Encode(id)), p2pEnv)
	assert.Error(t, err, "must return error")
	assert.Contains(t, err.Error(), "nil document provided")
	assert.Nil(t, resp, "must be nil")
}

func TestHandler_HandleInterceptor_getServiceAndModel_fail(t *testing.T) {
	ctx := testingconfig.CreateAccountContext(t, cfg)
	cd, err := documents.NewCoreDocumentWithCollaborators(nil, documents.CDTreePrefix)
	assert.NoError(t, err)
	req := &p2ppb.AnchorDocumentRequest{Document: &cd.Document}
	p2pEnv, err := p2pcommon.PrepareP2PEnvelope(ctx, cfg.GetNetworkID(), p2pcommon.MessageTypeSendAnchoredDoc, req)
	assert.NoError(t, err)

	id, _ := cfg.GetIdentityID()
	resp, err := handler.HandleInterceptor(context.Background(), defaultPID, protocol.ID(hexutil.Encode(id)), p2pEnv)
	assert.Error(t, err, "must return error")
	assert.Contains(t, err.Error(), "core document embed data is nil")
	assert.Nil(t, resp, "must be nil")
}

func TestP2PService_basicChecks(t *testing.T) {
	tests := []struct {
		header *p2ppb.Header
		err    error
	}{
		{
			header: &p2ppb.Header{NodeVersion: "someversion", NetworkIdentifier: 12},
			err:    errors.AppendError(version.IncompatibleVersionError("someversion"), incompatibleNetworkError(cfg.GetNetworkID(), 12)),
		},

		{
			header: &p2ppb.Header{NodeVersion: "0.0.1", NetworkIdentifier: 12},
			err:    errors.AppendError(incompatibleNetworkError(cfg.GetNetworkID(), 12), nil),
		},

		{
			header: &p2ppb.Header{NodeVersion: version.GetVersion().String(), NetworkIdentifier: cfg.GetNetworkID()},
		},
	}

	id, _ := cfg.GetIdentityID()
	centID := identity.NewDIDFromBytes(id)
	for _, c := range tests {
		err := HandshakeValidator(cfg.GetNetworkID(), mockIDService).Validate(c.header, &centID, &defaultPID)
		if err != nil {
			if c.err == nil {
				t.Fatalf("unexpected error: %v\n", err)
			}
			assert.EqualError(t, err, c.err.Error(), "error mismatch")
		}
	}

}
