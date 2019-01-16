// +build unit

package p2p

import (
	"context"
	"crypto/rand"
	"fmt"
	"io"
	"testing"
	"time"

	"github.com/centrifuge/centrifuge-protobufs/gen/go/p2p"
	"github.com/centrifuge/go-centrifuge/p2p/common"
	"github.com/centrifuge/go-centrifuge/testingutils/config"

	"github.com/centrifuge/go-centrifuge/errors"

	pb "github.com/centrifuge/go-centrifuge/protobufs/gen/go/protocol"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/libp2p/go-libp2p-crypto"
	"github.com/libp2p/go-libp2p-host"
	libp2pPeer "github.com/libp2p/go-libp2p-peer"
	"github.com/libp2p/go-libp2p-protocol"
	"github.com/stretchr/testify/assert"
)

const MessengerDummyProtocol protocol.ID = "/messeger/dummy/0.0.1"
const MessengerDummyProtocol2 protocol.ID = "/messeger/dummy/0.0.2"

var mockedHandler = func(ctx context.Context, peer libp2pPeer.ID, protoc protocol.ID, msg *pb.P2PEnvelope) (envelope *pb.P2PEnvelope, e error) {
	dataEnv, err := p2pcommon.ResolveDataEnvelope(msg)
	if err != nil {
		return nil, err
	}
	if p2pcommon.MessageTypeSendAnchoredDoc.Equals(dataEnv.Header.Type) {
		dataEnv.Header.Type = p2pcommon.MessageTypeSendAnchoredDocRep.String()
	} else if p2pcommon.MessageTypeRequestSignature.Equals(dataEnv.Header.Type) {
		dataEnv.Header.Type = p2pcommon.MessageTypeRequestSignatureRep.String()
	} else if p2pcommon.MessageTypeSendAnchoredDocRep.Equals(dataEnv.Header.Type) {
		dataEnv.Header.Type = p2pcommon.MessageTypeSendAnchoredDoc.String()
	} else if p2pcommon.MessageTypeError.Equals(dataEnv.Header.Type) {
		return nil, errors.New("dummy error")
	} else if p2pcommon.MessageTypeRequestSignatureRep.Equals(dataEnv.Header.Type) {
		return nil, nil
	} else if p2pcommon.MessageTypeInvalid.Equals(dataEnv.Header.Type) {
		return nil, errors.New("invalid data")
	}

	return p2pcommon.PrepareP2PEnvelope(ctx, uint32(0), p2pcommon.MessageTypeFromString(dataEnv.Header.Type), dataEnv)
}

// Using a single test for all cases to use the same hosts in a synchronous way
func TestHandleNewMessage(t *testing.T) {
	cfg, err := cfg.GetConfig()
	assert.NoError(t, err)
	cfg = updateKeys(cfg)
	ctx, canc := context.WithCancel(context.Background())
	c := testingconfig.CreateTenantContextWithContext(t, ctx, cfg)
	r := rand.Reader
	p1 := 35000
	p2 := 35001
	p3 := 35002
	h1 := createRandomHost(t, p1, r)
	h2 := createRandomHost(t, p2, r)
	h3 := createRandomHost(t, p3, r)
	// set h2 as the bootnode for h1
	_ = runDHT(c, h1, []string{fmt.Sprintf("/ip4/127.0.0.1/tcp/%d/ipfs/%s", p2, h2.ID().Pretty())})

	m1 := newP2PMessenger(c, h1, 5*time.Second, mockedHandler)
	m2 := newP2PMessenger(c, h2, 5*time.Second, mockedHandler)

	m1.init(MessengerDummyProtocol)
	m2.init(MessengerDummyProtocol)
	m2.init(MessengerDummyProtocol2)

	// 1. happy path
	// from h1 to h2 (with a message size ~ MessageSizeMax, has to be less because of the length bytes)
	p2pEnv, err := p2pcommon.PrepareP2PEnvelope(c, uint32(0), p2pcommon.MessageTypeRequestSignature, &p2ppb.Envelope{Body: utils.RandomSlice(MessageSizeMax - 400)})
	assert.NoError(t, err)
	msg, err := m1.sendMessage(c, h2.ID(), p2pEnv, MessengerDummyProtocol)
	assert.NoError(t, err)
	dataEnv, err := p2pcommon.ResolveDataEnvelope(msg)
	assert.NoError(t, err)
	assert.True(t, p2pcommon.MessageTypeRequestSignatureRep.Equals(dataEnv.Header.Type))

	// from h1 to h2 different protocol - intentionally setting reply-response in opposite for differentiation
	p2pEnv, err = p2pcommon.PrepareP2PEnvelope(c, uint32(0), p2pcommon.MessageTypeSendAnchoredDocRep, &p2ppb.Envelope{Body: utils.RandomSlice(3)})
	assert.NoError(t, err)
	msg, err = m1.sendMessage(c, h2.ID(), p2pEnv, MessengerDummyProtocol2)
	assert.NoError(t, err)
	dataEnv, err = p2pcommon.ResolveDataEnvelope(msg)
	assert.NoError(t, err)
	assert.True(t, p2pcommon.MessageTypeSendAnchoredDoc.Equals(dataEnv.Header.Type))

	// from h2 to h1
	p2pEnv, err = p2pcommon.PrepareP2PEnvelope(c, uint32(0), p2pcommon.MessageTypeSendAnchoredDoc, &p2ppb.Envelope{Body: utils.RandomSlice(3)})
	assert.NoError(t, err)
	msg, err = m2.sendMessage(c, h1.ID(), p2pEnv, MessengerDummyProtocol)
	assert.NoError(t, err)
	dataEnv, err = p2pcommon.ResolveDataEnvelope(msg)
	assert.NoError(t, err)
	assert.True(t, p2pcommon.MessageTypeSendAnchoredDocRep.Equals(dataEnv.Header.Type))

	// 2. unrecognized  protocol
	p2pEnv, err = p2pcommon.PrepareP2PEnvelope(c, uint32(0), p2pcommon.MessageTypeRequestSignature, &p2ppb.Envelope{Body: utils.RandomSlice(3)})
	assert.NoError(t, err)
	msg, err = m1.sendMessage(c, h2.ID(), p2pEnv, "wrong")
	if assert.Error(t, err) {
		assert.Equal(t, "protocol not supported", err.Error())
	}

	// 3. unrecognized message type - stream would be reset by the peer
	p2pEnv, err = p2pcommon.PrepareP2PEnvelope(c, uint32(0), p2pcommon.MessageTypeInvalid, &p2ppb.Envelope{Body: utils.RandomSlice(3)})
	assert.NoError(t, err)
	msg, err = m1.sendMessage(c, h2.ID(), p2pEnv, MessengerDummyProtocol)
	if assert.Error(t, err) {
		assert.Equal(t, "stream reset", err.Error())
	}

	// 4. handler error
	p2pEnv, err = p2pcommon.PrepareP2PEnvelope(c, uint32(0), p2pcommon.MessageTypeError, &p2ppb.Envelope{Body: utils.RandomSlice(3)})
	assert.NoError(t, err)
	msg, err = m1.sendMessage(c, h2.ID(), p2pEnv, MessengerDummyProtocol)
	if assert.Error(t, err) {
		assert.Equal(t, "stream reset", err.Error())
	}

	// 5. can't find host - h3
	p2pEnv, err = p2pcommon.PrepareP2PEnvelope(c, uint32(0), p2pcommon.MessageTypeRequestSignature, &p2ppb.Envelope{Body: utils.RandomSlice(3)})
	assert.NoError(t, err)
	msg, err = m1.sendMessage(c, h3.ID(), p2pEnv, MessengerDummyProtocol)
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "dial attempt failed: failed to dial <peer.ID")
	}

	// 6. handler nil response
	p2pEnv, err = p2pcommon.PrepareP2PEnvelope(c, uint32(0), p2pcommon.MessageTypeRequestSignatureRep, &p2ppb.Envelope{Body: utils.RandomSlice(3)})
	assert.NoError(t, err)
	msg, err = m1.sendMessage(c, h2.ID(), p2pEnv, MessengerDummyProtocol)
	if assert.Error(t, err) {
		assert.Equal(t, "timed out reading response", err.Error())
	}

	// 7. message size more than the max
	// from h1 to h2 (with a message size > MessageSizeMax)
	p2pEnv, err = p2pcommon.PrepareP2PEnvelope(c, uint32(0), p2pcommon.MessageTypeRequestSignature, &p2ppb.Envelope{Body: utils.RandomSlice(MessageSizeMax)})
	assert.NoError(t, err)
	msg, err = m1.sendMessage(c, h2.ID(), p2pEnv, MessengerDummyProtocol)
	if assert.Error(t, err) {
		assert.Equal(t, "stream reset", err.Error())
	}
	canc()
}

func createRandomHost(t *testing.T, port int, r io.Reader) host.Host {
	priv, pub, err := crypto.GenerateKeyPairWithReader(crypto.RSA, 2048, r)
	assert.NoError(t, err)
	h1, err := makeBasicHost(priv, pub, "", port)
	assert.NoError(t, err)
	return h1
}
