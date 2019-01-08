// +build unit

package p2p

import (
	"context"
	"crypto/rand"
	"fmt"
	"io"
	"testing"
	"time"

	"github.com/centrifuge/go-centrifuge/errors"

	pb "github.com/centrifuge/go-centrifuge/protobufs/gen/go/protocol"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/libp2p/go-libp2p-crypto"
	"github.com/libp2p/go-libp2p-host"
	"github.com/libp2p/go-libp2p-peer"
	"github.com/libp2p/go-libp2p-protocol"
	"github.com/stretchr/testify/assert"
)

const MessengerDummyProtocol protocol.ID = "/messeger/dummy/0.0.1"
const MessengerDummyProtocol2 protocol.ID = "/messeger/dummy/0.0.2"

// Using a single test for all cases to use the same hosts in a synchronous way
func TestHandleNewMessage(t *testing.T) {
	c, canc := context.WithCancel(context.Background())
	r := rand.Reader
	p1 := 35000
	p2 := 35001
	p3 := 35002
	h1 := createRandomHost(t, p1, r)
	h2 := createRandomHost(t, p2, r)
	h3 := createRandomHost(t, p3, r)
	// set h2 as the bootnode for h1
	_ = runDHT(c, h1, []string{fmt.Sprintf("/ip4/127.0.0.1/tcp/%d/ipfs/%s", p2, h2.ID().Pretty())})

	m1 := newP2PMessenger(c, h1, 1*time.Second)
	m1.addHandler(pb.MessageType_MESSAGE_TYPE_SEND_ANCHORED_DOC, func(ctx context.Context, peer peer.ID, protoc protocol.ID, msg *pb.P2PEnvelope) (envelope *pb.P2PEnvelope, e error) {
		return &pb.P2PEnvelope{Type: pb.MessageType_MESSAGE_TYPE_SEND_ANCHORED_DOC_REP}, nil
	})
	m2 := newP2PMessenger(c, h2, 1*time.Second)
	m2.addHandler(pb.MessageType_MESSAGE_TYPE_REQUEST_SIGNATURE, func(ctx context.Context, peer peer.ID, protoc protocol.ID, msg *pb.P2PEnvelope) (envelope *pb.P2PEnvelope, e error) {
		return &pb.P2PEnvelope{Type: pb.MessageType_MESSAGE_TYPE_REQUEST_SIGNATURE_REP}, nil
	})
	m2.addHandler(pb.MessageType_MESSAGE_TYPE_SEND_ANCHORED_DOC_REP, func(ctx context.Context, peer peer.ID, protoc protocol.ID, msg *pb.P2PEnvelope) (envelope *pb.P2PEnvelope, e error) {
		return &pb.P2PEnvelope{Type: pb.MessageType_MESSAGE_TYPE_SEND_ANCHORED_DOC}, nil
	})
	// error
	m2.addHandler(pb.MessageType_MESSAGE_TYPE_ERROR, func(ctx context.Context, peer peer.ID, protoc protocol.ID, msg *pb.P2PEnvelope) (envelope *pb.P2PEnvelope, e error) {
		return nil, errors.New("dummy error")
	})
	// nil response - message type here is irrelevant using the reply type for convenience
	m2.addHandler(pb.MessageType_MESSAGE_TYPE_REQUEST_SIGNATURE_REP, func(ctx context.Context, peer peer.ID, protoc protocol.ID, msg *pb.P2PEnvelope) (envelope *pb.P2PEnvelope, e error) {
		return nil, nil
	})

	m1.init(MessengerDummyProtocol)
	m2.init(MessengerDummyProtocol)
	m2.init(MessengerDummyProtocol2)

	// 1. happy path
	// from h1 to h2 (with a message size ~ MessageSizeMax, has to be less because of the length bytes)
	msg, err := m1.sendMessage(c, h2.ID(), &pb.P2PEnvelope{Type: pb.MessageType_MESSAGE_TYPE_REQUEST_SIGNATURE, Body: utils.RandomSlice(MessageSizeMax - 7)}, MessengerDummyProtocol)
	assert.NoError(t, err)
	assert.Equal(t, pb.MessageType_MESSAGE_TYPE_REQUEST_SIGNATURE_REP, msg.Type)
	// from h1 to h2 different protocol - intentionally setting reply-response in opposite for differentiation
	msg, err = m1.sendMessage(c, h2.ID(), &pb.P2PEnvelope{Type: pb.MessageType_MESSAGE_TYPE_SEND_ANCHORED_DOC_REP, Body: utils.RandomSlice(3)}, MessengerDummyProtocol2)
	assert.NoError(t, err)
	assert.Equal(t, pb.MessageType_MESSAGE_TYPE_SEND_ANCHORED_DOC, msg.Type)
	// from h2 to h1
	msg, err = m2.sendMessage(c, h1.ID(), &pb.P2PEnvelope{Type: pb.MessageType_MESSAGE_TYPE_SEND_ANCHORED_DOC, Body: utils.RandomSlice(3)}, MessengerDummyProtocol)
	assert.NoError(t, err)
	assert.Equal(t, pb.MessageType_MESSAGE_TYPE_SEND_ANCHORED_DOC_REP, msg.Type)

	// 2. unrecognized  protocol
	msg, err = m1.sendMessage(c, h2.ID(), &pb.P2PEnvelope{Type: pb.MessageType_MESSAGE_TYPE_REQUEST_SIGNATURE, Body: utils.RandomSlice(3)}, "wrong")
	if assert.Error(t, err) {
		assert.Equal(t, "protocol not supported", err.Error())
	}

	// 3. unrecognized message type - stream would be reset by the peer
	msg, err = m1.sendMessage(c, h2.ID(), &pb.P2PEnvelope{Type: pb.MessageType_MESSAGE_TYPE_INVALID, Body: utils.RandomSlice(3)}, MessengerDummyProtocol)
	if assert.Error(t, err) {
		assert.Equal(t, "stream reset", err.Error())
	}

	// 4. handler error
	msg, err = m1.sendMessage(c, h2.ID(), &pb.P2PEnvelope{Type: pb.MessageType_MESSAGE_TYPE_ERROR, Body: utils.RandomSlice(3)}, MessengerDummyProtocol)
	if assert.Error(t, err) {
		assert.Equal(t, "stream reset", err.Error())
	}

	// 5. can't find host - h3
	msg, err = m1.sendMessage(c, h3.ID(), &pb.P2PEnvelope{Type: pb.MessageType_MESSAGE_TYPE_REQUEST_SIGNATURE, Body: utils.RandomSlice(3)}, MessengerDummyProtocol)
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "dial attempt failed: failed to dial <peer.ID")
	}

	// 6. handler nil response
	msg, err = m1.sendMessage(c, h2.ID(), &pb.P2PEnvelope{Type: pb.MessageType_MESSAGE_TYPE_REQUEST_SIGNATURE_REP, Body: utils.RandomSlice(3)}, MessengerDummyProtocol)
	if assert.Error(t, err) {
		assert.Equal(t, "timed out reading response", err.Error())
	}

	// 7. message size more than the max
	// from h1 to h2 (with a message size > MessageSizeMax)
	msg, err = m1.sendMessage(c, h2.ID(), &pb.P2PEnvelope{Type: pb.MessageType_MESSAGE_TYPE_REQUEST_SIGNATURE, Body: utils.RandomSlice(MessageSizeMax)}, MessengerDummyProtocol)
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
