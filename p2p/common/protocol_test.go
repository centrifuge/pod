//go:build unit
// +build unit

package p2pcommon

import (
	"os"
	"testing"

	p2ppb "github.com/centrifuge/centrifuge-protobufs/gen/go/p2p"
	protocolpb "github.com/centrifuge/centrifuge-protobufs/gen/go/protocol"
	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/bootstrap/bootstrappers/testlogging"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/golang/protobuf/proto"
	"github.com/libp2p/go-libp2p-core/protocol"
	"github.com/stretchr/testify/assert"
)

var cfg config.Configuration

func TestMain(m *testing.M) {
	ibootstappers := []bootstrap.TestBootstrapper{
		&testlogging.TestLoggingBootstrapper{},
		&config.Bootstrapper{},
	}
	ctx := make(map[string]interface{})
	bootstrap.RunTestBootstrappers(ibootstappers, ctx)
	cfg = ctx[bootstrap.BootstrappedConfig].(config.Configuration)
	result := m.Run()
	bootstrap.RunTestTeardown(ibootstappers)
	os.Exit(result)
}

func TestExtractDID(t *testing.T) {
	p := protocol.ID("/centrifuge/0.0.1/0xBAEb33a61f05e6F269f1c4b4CFF91A901B54DaF7")
	cid, err := ExtractIdentity(p)
	assert.NoError(t, err)
	assert.Equal(t, "0xBAEb33a61f05e6F269f1c4b4CFF91A901B54DaF7", cid.String())
}

func TestProtocolForCID(t *testing.T) {
	cid, err := identity.NewDIDFromString("0xBAEb33a61f05e6F269f1c4b4CFF91A901B54DaF7")
	assert.NoError(t, err)
	p := ProtocolForIdentity(cid)
	assert.Contains(t, p, cid.String())
	cidE, err := ExtractIdentity(p)
	assert.NoError(t, err)
	assert.NoError(t, err)
	assert.Equal(t, cid.String(), cidE.String())
}

func TestResolveDataEnvelope(t *testing.T) {
	// Nil Payload
	dataEnv, err := ResolveDataEnvelope(nil)
	assert.Nil(t, dataEnv)
	assert.Error(t, err, "must return error")
	assert.Contains(t, err.Error(), "cannot cast proto.Message to protocolpb.P2PEnvelope")

	// Missing Header field
	body, err := proto.Marshal(&p2ppb.Envelope{})
	msg := &protocolpb.P2PEnvelope{
		Body: body,
	}
	dataEnv, err = ResolveDataEnvelope(msg)
	assert.Nil(t, dataEnv)
	assert.Error(t, err, "must return error")
	assert.Contains(t, err.Error(), "Header field is empty")

	// Success
	body, err = proto.Marshal(&p2ppb.Envelope{Header: &p2ppb.Header{Type: "messageType"}})
	msg = &protocolpb.P2PEnvelope{
		Body: body,
	}
	dataEnv, err = ResolveDataEnvelope(msg)
	assert.Nil(t, err)
	assert.Equal(t, "messageType", dataEnv.Header.Type)
}

func TestPrepareP2PEnvelope(t *testing.T) {
	//// Missing Self
	//p2pEnv, err := PrepareP2PEnvelope(context.Background(), uint32(0), MessageTypeRequestSignature, nil)
	//assert.Error(t, err)
	//assert.Nil(t, p2pEnv)
	//
	//acc := &configstore.Account{
	//	IdentityID: id,
	//	P2PKeyPair: configstore.KeyPair{
	//		Pvt: ssk,
	//		Pub: spk,
	//	},
	//	SigningKeyPair: configstore.KeyPair{
	//		Pvt: ssk,
	//		Pub: spk,
	//	},
	//}
	//ctx, _ := contextutil.New(context.Background(), acc)
	//assert.NotNil(t, ctx)
	//
	//// Nil proto.Message
	//p2pEnv, err = PrepareP2PEnvelope(ctx, uint32(0), MessageTypeRequestSignature, nil)
	//assert.Error(t, err)
	//
	//// Success
	//msg := &protocolpb.P2PEnvelope{Body: utils.RandomSlice(3)}
	//p2pEnv, err = PrepareP2PEnvelope(ctx, uint32(0), MessageTypeRequestSignature, msg)
	//assert.NoError(t, err)
	//assert.NotNil(t, p2pEnv)
	//dataEnv, err := ResolveDataEnvelope(p2pEnv)
	//assert.NoError(t, err)
	//assert.NotNil(t, dataEnv)
}
