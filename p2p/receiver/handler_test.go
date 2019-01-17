// +build unit

package receiver

import (
	"context"
	"crypto/rand"
	"os"
	"testing"

	"github.com/centrifuge/go-centrifuge/documents/genericdoc"

	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/p2p/common"
	"github.com/centrifuge/go-centrifuge/protobufs/gen/go/protocol"
	testingcommons "github.com/centrifuge/go-centrifuge/testingutils/commons"
	"github.com/centrifuge/go-centrifuge/testingutils/config"
	"github.com/centrifuge/go-centrifuge/version"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/golang/protobuf/proto"
	"github.com/libp2p/go-libp2p-protocol"

	"github.com/centrifuge/go-centrifuge/storage/leveldb"

	"github.com/centrifuge/go-centrifuge/config/configstore"

	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/centrifuge-protobufs/gen/go/p2p"
	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/bootstrap/bootstrappers/testlogging"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/coredocument"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/queue"
	"github.com/centrifuge/go-centrifuge/testingutils/coredocument"
	"github.com/centrifuge/go-centrifuge/transactions"
	"github.com/golang/protobuf/ptypes/any"
	"github.com/libp2p/go-libp2p-crypto"
	libp2pPeer "github.com/libp2p/go-libp2p-peer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var (
	handler       *Handler
	registry      *documents.ServiceRegistry
	coreDoc       = testingcoredocument.GenerateCoreDocument()
	cfg           config.Configuration
	mockIDService *testingcommons.MockIDService
	defaultPID    libp2pPeer.ID
)

func TestMain(m *testing.M) {
	ibootstappers := []bootstrap.TestBootstrapper{
		&testlogging.TestLoggingBootstrapper{},
		&config.Bootstrapper{},
		&leveldb.Bootstrapper{},
		&configstore.Bootstrapper{},
		&queue.Bootstrapper{},
		transactions.Bootstrapper{},
		documents.Bootstrapper{},
	}
	ctx := make(map[string]interface{})
	ctx[identity.BootstrappedIDService] = &testingcommons.MockIDService{}
	bootstrap.RunTestBootstrappers(ibootstappers, ctx)
	cfg = ctx[bootstrap.BootstrappedConfig].(config.Configuration)
	cfgService := ctx[config.BootstrappedConfigStorage].(config.Service)
	registry = ctx[documents.BootstrappedRegistry].(*documents.ServiceRegistry)
	genService := genericdoc.DefaultService(nil, nil, nil)
	mockIDService = &testingcommons.MockIDService{}
	_, pub, _ := crypto.GenerateEd25519Key(rand.Reader)
	defaultPID, _ = libp2pPeer.IDFromPublicKey(pub)
	mockIDService.On("ValidateKey", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	handler = New(cfgService, registry, HandshakeValidator(cfg.GetNetworkID(), mockIDService), genService)
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
	ctx := testingconfig.CreateTenantContext(t, cfg)
	p2pEnv, err := p2pcommon.PrepareP2PEnvelope(ctx, cfg.GetNetworkID(), p2pcommon.MessageTypeRequestSignature, &protocolpb.P2PEnvelope{})
	assert.NoError(t, err)
	resp, err := handler.HandleInterceptor(context.Background(), libp2pPeer.ID("SomePeer"), protocol.ID("protocolX"), p2pEnv)
	assert.Error(t, err, "must return error")
	assert.Contains(t, err.Error(), "hex string without 0x prefix")
	assert.Nil(t, resp, "must be nil")
}

func TestHandler_HandleInterceptor_TenantNotFound(t *testing.T) {
	ctx := testingconfig.CreateTenantContext(t, cfg)
	p2pEnv, err := p2pcommon.PrepareP2PEnvelope(ctx, cfg.GetNetworkID(), p2pcommon.MessageTypeRequestSignature, &protocolpb.P2PEnvelope{})
	assert.NoError(t, err)
	resp, err := handler.HandleInterceptor(context.Background(), libp2pPeer.ID("SomePeer"), protocol.ID("0x001100110011"), p2pEnv)
	assert.Error(t, err, "must return error")
	assert.Contains(t, err.Error(), "model not found in db")
	assert.Nil(t, resp, "must be nil")
}

func TestHandler_HandleInterceptor_HandshakeValidationFail(t *testing.T) {
	ctx := testingconfig.CreateTenantContext(t, cfg)
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
	ctx := testingconfig.CreateTenantContext(t, cfg)
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
	ctx := testingconfig.CreateTenantContext(t, cfg)
	p2pEnv, err := p2pcommon.PrepareP2PEnvelope(ctx, cfg.GetNetworkID(), p2pcommon.MessageTypeRequestSignature, &protocolpb.P2PEnvelope{})
	assert.NoError(t, err)

	id, _ := cfg.GetIdentityID()
	resp, err := handler.HandleInterceptor(context.Background(), defaultPID, protocol.ID(hexutil.Encode(id)), p2pEnv)
	assert.Error(t, err, "must return error")
	assert.Contains(t, err.Error(), "nil core document")
	assert.Nil(t, resp, "must be nil")
}

func TestHandler_HandleInterceptor_getServiceAndModel_fail(t *testing.T) {
	ctx := testingconfig.CreateTenantContext(t, cfg)
	req := &p2ppb.AnchorDocumentRequest{Document: coredocument.New()}
	p2pEnv, err := p2pcommon.PrepareP2PEnvelope(ctx, cfg.GetNetworkID(), p2pcommon.MessageTypeSendAnchoredDoc, req)
	assert.NoError(t, err)

	id, _ := cfg.GetIdentityID()
	resp, err := handler.HandleInterceptor(context.Background(), defaultPID, protocol.ID(hexutil.Encode(id)), p2pEnv)
	assert.Error(t, err, "must return error")
	assert.Contains(t, err.Error(), "failed to get type of the document")
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
	centID, _ := identity.ToCentID(id)
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

type mockModel struct {
	mock.Mock
	documents.Model
}

type mockService struct {
	mock.Mock
	documents.Service
}

func (s mockService) DeriveFromCoreDocument(cd *coredocumentpb.CoreDocument) (documents.Model, error) {
	args := s.Called(cd)
	m, _ := args.Get(0).(documents.Model)
	return m, args.Error(1)
}

func Test_getServiceAndModel(t *testing.T) {
	// document nil fail
	s, m, err := getServiceAndModel(registry, nil)
	assert.Error(t, err)

	// docType fetch fail
	cd := coredocument.New()
	s, m, err = getServiceAndModel(registry, cd)
	assert.Error(t, err)
	assert.Nil(t, s)
	assert.Nil(t, m)
	assert.Contains(t, err.Error(), "failed to get type of the document")

	// missing service
	cd.EmbeddedData = &any.Any{
		TypeUrl: "model_type_fail",
		Value:   []byte("some data"),
	}
	s, m, err = getServiceAndModel(registry, cd)
	assert.Error(t, err)
	assert.Nil(t, s)
	assert.Nil(t, m)
	assert.Contains(t, err.Error(), "failed to locate the service")

	// derive fails
	srv := mockService{}
	srv.On("DeriveFromCoreDocument", cd).Return(nil, errors.New("error")).Once()
	err = registry.Register(cd.EmbeddedData.TypeUrl, srv)
	assert.Nil(t, err)
	s, m, err = getServiceAndModel(registry, cd)
	srv.AssertExpectations(t)
	assert.Error(t, err)
	assert.Nil(t, s)
	assert.Nil(t, m)
	assert.Contains(t, err.Error(), "failed to derive model from core document")

	// success
	model := &mockModel{}
	cd.EmbeddedData.TypeUrl = "get_model_type"
	srv = mockService{}
	srv.On("DeriveFromCoreDocument", cd).Return(model, nil).Once()
	err = registry.Register(cd.EmbeddedData.TypeUrl, srv)
	assert.Nil(t, err)
	s, m, err = getServiceAndModel(registry, cd)
	srv.AssertExpectations(t)
	assert.Nil(t, err)
	assert.NotNil(t, s)
	assert.NotNil(t, m)
	assert.Equal(t, model, m)
}
