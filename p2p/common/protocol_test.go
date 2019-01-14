// +build unit

package p2pcommon

import (
	"context"
	"os"
	"testing"

	"github.com/centrifuge/centrifuge-protobufs/gen/go/p2p"
	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/bootstrap/bootstrappers/testlogging"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/config/configstore"
	"github.com/centrifuge/go-centrifuge/contextutil"
	"github.com/centrifuge/go-centrifuge/protobufs/gen/go/protocol"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/golang/protobuf/proto"

	"github.com/centrifuge/go-centrifuge/identity"

	"github.com/libp2p/go-libp2p-protocol"
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

func TestExtractCID(t *testing.T) {
	p := protocol.ID("/centrifuge/0.0.1/0xd9f72e705074")
	cid, err := ExtractCID(p)
	assert.NoError(t, err)
	assert.Equal(t, "0xd9f72e705074", cid.String())
}

func TestProtocolForCID(t *testing.T) {
	cid := identity.RandomCentID()
	p := ProtocolForCID(cid)
	assert.Contains(t, p, cid.String())
	cidE, err := ExtractCID(p)
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
	// Missing Self
	p2pEnv, err := PrepareP2PEnvelope(context.Background(), uint32(0), MessageTypeRequestSignature, nil)
	assert.Error(t, err)
	assert.Nil(t, p2pEnv)

	id, _ := cfg.GetIdentityID()
	spk, ssk := cfg.GetSigningKeyPair()
	tc := &configstore.TenantConfig{
		IdentityID: id,
		SigningKeyPair: configstore.KeyPair{
			Priv: ssk,
			Pub:  spk,
		},
		EthAuthKeyPair: configstore.KeyPair{
			Priv: ssk,
			Pub:  spk,
		},
	}
	ctx, _ := contextutil.NewCentrifugeContext(context.Background(), tc)
	assert.NotNil(t, ctx)

	// Nil proto.Message
	p2pEnv, err = PrepareP2PEnvelope(ctx, uint32(0), MessageTypeRequestSignature, nil)
	assert.Error(t, err)

	// Success
	msg := &protocolpb.P2PEnvelope{Body: utils.RandomSlice(3)}
	p2pEnv, err = PrepareP2PEnvelope(ctx, uint32(0), MessageTypeRequestSignature, msg)
	assert.NoError(t, err)
	assert.NotNil(t, p2pEnv)
	dataEnv, err := ResolveDataEnvelope(p2pEnv)
	assert.NoError(t, err)
	assert.NotNil(t, dataEnv.Header.Signature)
}
