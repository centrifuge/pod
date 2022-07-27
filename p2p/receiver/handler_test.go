//go:build unit

package receiver

import (
	"context"
	"crypto/rand"
	"os"
	"testing"
	"time"

	errorspb "github.com/centrifuge/centrifuge-protobufs/gen/go/errors"
	p2ppb "github.com/centrifuge/centrifuge-protobufs/gen/go/p2p"
	protocolpb "github.com/centrifuge/centrifuge-protobufs/gen/go/protocol"
	"github.com/centrifuge/go-centrifuge/anchors"
	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/bootstrap/bootstrappers/testlogging"
	"github.com/centrifuge/go-centrifuge/centchain"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/config/configstore"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/ethereum"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/jobs"
	p2pcommon "github.com/centrifuge/go-centrifuge/p2p/common"
	"github.com/centrifuge/go-centrifuge/storage/leveldb"
	testingcommons "github.com/centrifuge/go-centrifuge/testingutils/commons"
	testingconfig "github.com/centrifuge/go-centrifuge/testingutils/config"
	testingdocuments "github.com/centrifuge/go-centrifuge/testingutils/documents"
	testingidentity "github.com/centrifuge/go-centrifuge/testingutils/identity"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/centrifuge/go-centrifuge/version"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/golang/protobuf/proto"
	"github.com/libp2p/go-libp2p-core/crypto"
	libp2pPeer "github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/protocol"
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
	ethClient := &ethereum.MockEthClient{}
	ethClient.On("GetEthClient").Return(nil)
	ctx[ethereum.BootstrappedEthereumClient] = ethClient
	centChainClient := &centchain.MockAPI{}
	ctx[centchain.BootstrappedCentChainClient] = centChainClient
	ibootstappers := []bootstrap.TestBootstrapper{
		&testlogging.TestLoggingBootstrapper{},
		&config.Bootstrapper{},
		&leveldb.Bootstrapper{},
		jobs.Bootstrapper{},
		&configstore.Bootstrapper{},
		&anchors.Bootstrapper{},
		documents.Bootstrapper{},
	}
	errors.MaskErrs = false
	mockIDService = &testingcommons.MockIdentityService{}
	ctx[identity.BootstrappedDIDService] = mockIDService
	ctx[identity.BootstrappedDIDFactory] = &identity.MockFactory{}
	bootstrap.RunTestBootstrappers(ibootstappers, ctx)
	cfg = ctx[bootstrap.BootstrappedConfig].(config.Configuration)
	cfgService := ctx[config.BootstrappedConfigStorage].(config.Service)
	registry = ctx[documents.BootstrappedRegistry].(*documents.ServiceRegistry)
	docSrv := documents.DefaultService(cfg, nil, nil, registry, mockIDService, nil)
	_, pub, _ := crypto.GenerateEd25519Key(rand.Reader)
	defaultPID, _ = libp2pPeer.IDFromPublicKey(pub)
	mockIDService.On("ValidateKey", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	handler = New(cfgService, HandshakeValidator(cfg.GetNetworkID(), mockIDService), docSrv, new(testingdocuments.MockRegistry), mockIDService)
	result := m.Run()
	bootstrap.RunTestTeardown(ibootstappers)
	os.Exit(result)
}

func TestHandler_HandleInterceptor_noConfig(t *testing.T) {
	randomPath := leveldb.GetRandomTestStoragePath()
	defer os.RemoveAll(randomPath)
	db, err := leveldb.NewLevelDBStorage(randomPath)
	assert.NoError(t, err)
	fkRepo := configstore.NewDBRepository(leveldb.NewLevelDBRepository(db))
	fkCfg := configstore.DefaultService(fkRepo, mockIDService)
	hndlr := New(fkCfg, nil, nil, nil, nil)
	resp, err := hndlr.HandleInterceptor(context.Background(), libp2pPeer.ID("SomePeer"), protocol.ID("protocolX"), &protocolpb.P2PEnvelope{})
	assert.NoError(t, err)
	err = p2pcommon.ConvertP2PEnvelopeToError(resp)
	assert.Error(t, err)
}

func TestHandler_RequestDocumentSignature_nilDocument(t *testing.T) {
	id := testingidentity.GenerateRandomDID()
	req := &p2ppb.SignatureRequest{}
	resp, err := handler.RequestDocumentSignature(context.Background(), req, id)
	assert.Error(t, err, "must return error")
	assert.Nil(t, resp, "must be nil")
}

func TestHandler_HandleInterceptor_nilPayload(t *testing.T) {
	resp, err := handler.HandleInterceptor(context.Background(), libp2pPeer.ID("SomePeer"), protocol.ID("protocolX"), nil)
	assert.NoError(t, err)
	err = p2pcommon.ConvertP2PEnvelopeToError(resp)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "nil payload provided")
}

func TestHandler_HandleInterceptor_HeaderEmpty(t *testing.T) {
	resp, err := handler.HandleInterceptor(context.Background(), libp2pPeer.ID("SomePeer"), protocol.ID("protocolX"), &protocolpb.P2PEnvelope{})
	assert.NoError(t, err)
	err = p2pcommon.ConvertP2PEnvelopeToError(resp)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Header field is empty")
}

func TestHandler_HandleInterceptor_CentIDNotHex(t *testing.T) {
	ctx := testingconfig.CreateAccountContext(t, cfg)
	p2pEnv, err := p2pcommon.PrepareP2PEnvelope(ctx, cfg.GetNetworkID(), p2pcommon.MessageTypeRequestSignature, &protocolpb.P2PEnvelope{})
	assert.NoError(t, err)
	resp, err := handler.HandleInterceptor(context.Background(), libp2pPeer.ID("SomePeer"), protocol.ID("protocolX"), p2pEnv)
	assert.NoError(t, err)
	err = p2pcommon.ConvertP2PEnvelopeToError(resp)
	assert.Error(t, err)
	assert.Equal(t, identity.ErrMalformedAddress.Error(), err.Error())
}

func TestHandler_HandleInterceptor_TenantNotFound(t *testing.T) {
	ctx := testingconfig.CreateAccountContext(t, cfg)
	p2pEnv, err := p2pcommon.PrepareP2PEnvelope(ctx, cfg.GetNetworkID(), p2pcommon.MessageTypeRequestSignature, &protocolpb.P2PEnvelope{})
	assert.NoError(t, err)
	resp, err := handler.HandleInterceptor(context.Background(), libp2pPeer.ID("SomePeer"), protocol.ID("0x89b0a86583c4442acfd71b463e0d3c55ae1412a5"), p2pEnv)
	assert.NoError(t, err)
	err = p2pcommon.ConvertP2PEnvelopeToError(resp)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "model not found in db")
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
	assert.NoError(t, err)
	err = p2pcommon.ConvertP2PEnvelopeToError(resp)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Incompatible version")

	// Manipulate network in Header
	p2pEnv, err = p2pcommon.PrepareP2PEnvelope(ctx, uint32(999), p2pcommon.MessageTypeRequestSignature, &protocolpb.P2PEnvelope{})
	assert.NoError(t, err)

	resp, err = handler.HandleInterceptor(context.Background(), libp2pPeer.ID("SomePeer"), protocol.ID(hexutil.Encode(id)), p2pEnv)
	assert.NoError(t, err)
	err = p2pcommon.ConvertP2PEnvelopeToError(resp)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Incompatible network id")
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
	assert.NoError(t, err)
	err = p2pcommon.ConvertP2PEnvelopeToError(resp)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "MessageType [UnsupportedType] not found")
}

func TestHandler_HandleInterceptor_NilDocument(t *testing.T) {
	ctx := testingconfig.CreateAccountContext(t, cfg)
	p2pEnv, err := p2pcommon.PrepareP2PEnvelope(ctx, cfg.GetNetworkID(), p2pcommon.MessageTypeRequestSignature, &protocolpb.P2PEnvelope{})
	assert.NoError(t, err)

	id, _ := cfg.GetIdentityID()
	resp, err := handler.HandleInterceptor(context.Background(), defaultPID, protocol.ID(hexutil.Encode(id)), p2pEnv)
	assert.NoError(t, err)
	err = p2pcommon.ConvertP2PEnvelopeToError(resp)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "nil document provided")
}

func TestHandler_HandleInterceptor_getServiceAndModel_fail(t *testing.T) {
	ctx := testingconfig.CreateAccountContext(t, cfg)
	cd, err := documents.NewCoreDocument(nil, documents.CollaboratorsAccess{}, nil)
	assert.NoError(t, err)
	req := &p2ppb.AnchorDocumentRequest{Document: cd.GetTestCoreDocWithReset()}
	p2pEnv, err := p2pcommon.PrepareP2PEnvelope(ctx, cfg.GetNetworkID(), p2pcommon.MessageTypeSendAnchoredDoc, req)
	assert.NoError(t, err)

	id, _ := cfg.GetIdentityID()
	resp, err := handler.HandleInterceptor(context.Background(), defaultPID, protocol.ID(hexutil.Encode(id)), p2pEnv)
	assert.NoError(t, err)
	err = p2pcommon.ConvertP2PEnvelopeToError(resp)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "core document embed data is nil")
}

func TestP2PService_basicChecks(t *testing.T) {
	tm, err := utils.ToTimestamp(time.Now())
	assert.NoError(t, err)
	tests := []struct {
		header *p2ppb.Header
		err    error
	}{
		{
			header: &p2ppb.Header{NodeVersion: "someversion", NetworkIdentifier: 12, Timestamp: tm},
			err:    errors.AppendError(version.IncompatibleVersionError("someversion"), incompatibleNetworkError(cfg.GetNetworkID(), 12)),
		},

		{
			header: &p2ppb.Header{NodeVersion: "2.0.0", NetworkIdentifier: 12, Timestamp: tm},
			err:    errors.AppendError(incompatibleNetworkError(cfg.GetNetworkID(), 12), nil),
		},

		{
			header: &p2ppb.Header{NodeVersion: version.GetVersion().String(), NetworkIdentifier: cfg.GetNetworkID(), Timestamp: tm},
		},
	}

	id, _ := cfg.GetIdentityID()
	centID, err := identity.NewDIDFromBytes(id)
	assert.NoError(t, err)
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

func TestConvertToErrorEnvelope(t *testing.T) {
	errPayload := errors.New("Error for P2P")
	envelope, err := handler.convertToErrorEnvelop(errPayload)
	assert.NoError(t, err)
	assert.NotNil(t, envelope)
	env, err := p2pcommon.ResolveDataEnvelope(envelope)
	assert.NoError(t, err)
	assert.Equal(t, p2pcommon.MessageTypeError.String(), env.Header.Type)

	// Unmarshal error PB
	m := new(errorspb.Error)
	errx := proto.Unmarshal(env.Body, m)
	assert.NoError(t, errx)
	assert.Equal(t, errPayload.Error(), m.Message)
}
