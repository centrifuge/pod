//go:build unit

package messenger

import (
	"bufio"
	"bytes"
	"context"
	"encoding/binary"
	"io"
	"testing"
	"time"

	protocolpb "github.com/centrifuge/centrifuge-protobufs/gen/go/protocol"
	"github.com/centrifuge/go-centrifuge/errors"
	p2pMocks "github.com/centrifuge/go-centrifuge/p2p/mocks"
	"github.com/centrifuge/go-centrifuge/p2p/receiver"
	genericUtils "github.com/centrifuge/go-centrifuge/testingutils/generic"
	"github.com/centrifuge/go-centrifuge/utils"
	inet "github.com/libp2p/go-libp2p-core/network"
	libp2ppeer "github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/protocol"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/protobuf/proto"
)

func TestP2PMessenger_Init(t *testing.T) {
	mes, mocks := getMessengerWithMocks(t)

	protocolID1 := protocol.ID("one")
	protocolID2 := protocol.ID("two")

	genericUtils.GetMock[*p2pMocks.HostMock](mocks).
		On(
			"SetStreamHandler",
			protocolID1,
			mock.AnythingOfType("network.StreamHandler"),
		).Once()

	genericUtils.GetMock[*p2pMocks.HostMock](mocks).
		On(
			"SetStreamHandler",
			protocolID2,
			mock.AnythingOfType("network.StreamHandler"),
		).Once()

	mes.Init(protocolID1, protocolID2)
}

func TestP2PMessenger_SendMessage(t *testing.T) {
	mes, mocks := getMessengerWithMocks(t)

	ctx := context.Background()

	peerID := libp2ppeer.ID("peer-id")
	protocolID := protocol.ID("protocol-id")

	msg := &protocolpb.P2PEnvelope{}

	senderMock := NewMessageSenderMock(t)

	args := &MessageSenderArgs{
		Ctx:        mes.ctx,
		Host:       mes.host,
		Timeout:    mes.timeout,
		PeerID:     peerID,
		ProtocolID: protocolID,
	}

	genericUtils.GetMock[*MessageSenderFactoryMock](mocks).
		On(
			"NewMessageSender",
			args,
		).Return(senderMock).Once()

	senderMock.On("Prepare").
		Return(nil).
		Once()

	senderRes := &protocolpb.P2PEnvelope{}

	senderMock.On("SendMessage", ctx, msg).
		Return(senderRes, nil).Once()

	res, err := mes.SendMessage(ctx, peerID, msg, protocolID)
	assert.NoError(t, err)
	assert.Equal(t, senderRes, res)
}

func TestP2PMessenger_SendMessage_SenderPrepareError(t *testing.T) {
	mes, mocks := getMessengerWithMocks(t)

	ctx := context.Background()

	peerID := libp2ppeer.ID("peer-id")
	protocolID := protocol.ID("protocol-id")

	msg := &protocolpb.P2PEnvelope{}

	senderMock := NewMessageSenderMock(t)

	args := &MessageSenderArgs{
		Ctx:        mes.ctx,
		Host:       mes.host,
		Timeout:    mes.timeout,
		PeerID:     peerID,
		ProtocolID: protocolID,
	}

	genericUtils.GetMock[*MessageSenderFactoryMock](mocks).
		On(
			"NewMessageSender",
			args,
		).Return(senderMock).Once()

	senderErr := errors.New("error")

	senderMock.On("Prepare").
		Return(senderErr).
		Once()

	res, err := mes.SendMessage(ctx, peerID, msg, protocolID)
	assert.ErrorIs(t, err, senderErr)
	assert.Nil(t, res)
}

func TestP2PMessenger_SendMessage_SendMessageError(t *testing.T) {
	mes, mocks := getMessengerWithMocks(t)

	ctx := context.Background()

	peerID := libp2ppeer.ID("peer-id")
	protocolID := protocol.ID("protocol-id")

	msg := &protocolpb.P2PEnvelope{}

	senderMock := NewMessageSenderMock(t)

	args := &MessageSenderArgs{
		Ctx:        mes.ctx,
		Host:       mes.host,
		Timeout:    mes.timeout,
		PeerID:     peerID,
		ProtocolID: protocolID,
	}

	genericUtils.GetMock[*MessageSenderFactoryMock](mocks).
		On(
			"NewMessageSender",
			args,
		).Return(senderMock).Once()

	senderMock.On("Prepare").
		Return(nil).
		Once()

	senderErr := errors.New("error")

	senderMock.On("SendMessage", ctx, msg).
		Return(nil, senderErr).Once()

	res, err := mes.SendMessage(ctx, peerID, msg, protocolID)
	assert.ErrorIs(t, err, senderErr)
	assert.Nil(t, res)
}

func TestP2PMessenger_handleNewMessage(t *testing.T) {
	mes, mocks := getMessengerWithMocks(t)

	requestEnvelope := &protocolpb.P2PEnvelope{
		Body: utils.RandomSlice(32),
	}

	requestEnvelopeBytes := getEncodedEnvelopeBytes(t, requestEnvelope)

	bufBytes := make([]byte, 0, 4096)
	buf := bytes.NewBuffer(bufBytes)

	connMock := p2pMocks.NewConnMock(t)

	peerID := libp2ppeer.ID("peer-id")

	connMock.On("RemotePeer").
		Return(peerID).Once()

	protocolID := protocol.ID("protocol-id")

	testStream := &testStream{
		r:          bytes.NewReader(requestEnvelopeBytes),
		w:          buf,
		conn:       connMock,
		protocolID: protocolID,
	}

	responseEnvelope := &protocolpb.P2PEnvelope{
		Body: utils.RandomSlice(32),
	}

	genericUtils.GetMock[*receiver.HandlerMock](mocks).
		On(
			"HandleInterceptor",
			mock.Anything,
			peerID,
			protocolID,
			mock.IsType(&protocolpb.P2PEnvelope{}),
		).Run(
		func(args mock.Arguments) {
			env := args.Get(3).(*protocolpb.P2PEnvelope)

			assert.Equal(t, requestEnvelope.GetBody(), env.GetBody())
		}).
		Return(responseEnvelope, nil).Once()

	mes.handleNewMessage(testStream)

	assertReaderContainsEnvelopeWithBody(t, buf, responseEnvelope.GetBody())
}

func TestP2PMessenger_handleNewMessage_ReadError(t *testing.T) {
	mes, _ := getMessengerWithMocks(t)

	bufBytes := make([]byte, 0, 4096)
	buf := bytes.NewBuffer(bufBytes)

	connMock := p2pMocks.NewConnMock(t)

	peerID := libp2ppeer.ID("peer-id")

	connMock.On("RemotePeer").
		Return(peerID).Once()

	protocolID := protocol.ID("protocol-id")

	testStream := &testStream{
		r:          bytes.NewReader(nil),
		w:          buf,
		conn:       connMock,
		protocolID: protocolID,
	}

	mes.handleNewMessage(testStream)

	assert.True(t, testStream.wasReset)
	assert.Len(t, buf.Bytes(), 0)
}

func TestP2PMessenger_handleNewMessage_NilHandler(t *testing.T) {
	mes, _ := getMessengerWithMocks(t)

	mes.handler = nil

	requestEnvelope := &protocolpb.P2PEnvelope{
		Body: utils.RandomSlice(32),
	}

	requestEnvelopeBytes := getEncodedEnvelopeBytes(t, requestEnvelope)

	bufBytes := make([]byte, 0, 4096)
	buf := bytes.NewBuffer(bufBytes)

	connMock := p2pMocks.NewConnMock(t)

	peerID := libp2ppeer.ID("peer-id")

	connMock.On("RemotePeer").
		Return(peerID).Once()

	protocolID := protocol.ID("protocol-id")

	testStream := &testStream{
		r:          bytes.NewReader(requestEnvelopeBytes),
		w:          buf,
		conn:       connMock,
		protocolID: protocolID,
	}

	mes.handleNewMessage(testStream)

	assert.True(t, testStream.wasReset)
	assert.Len(t, buf.Bytes(), 0)
}

func TestP2PMessenger_handleNewMessage_HandlerError(t *testing.T) {
	mes, mocks := getMessengerWithMocks(t)

	requestEnvelope := &protocolpb.P2PEnvelope{
		Body: utils.RandomSlice(32),
	}

	requestEnvelopeBytes := getEncodedEnvelopeBytes(t, requestEnvelope)

	bufBytes := make([]byte, 0, 4096)
	buf := bytes.NewBuffer(bufBytes)

	connMock := p2pMocks.NewConnMock(t)

	peerID := libp2ppeer.ID("peer-id")

	connMock.On("RemotePeer").
		Return(peerID).Once()

	protocolID := protocol.ID("protocol-id")

	testStream := &testStream{
		r:          bytes.NewReader(requestEnvelopeBytes),
		w:          buf,
		conn:       connMock,
		protocolID: protocolID,
	}

	genericUtils.GetMock[*receiver.HandlerMock](mocks).
		On(
			"HandleInterceptor",
			mock.Anything,
			peerID,
			protocolID,
			mock.IsType(&protocolpb.P2PEnvelope{}),
		).Run(
		func(args mock.Arguments) {
			env := args.Get(3).(*protocolpb.P2PEnvelope)

			assert.Equal(t, requestEnvelope.GetBody(), env.GetBody())
		}).
		Return(nil, errors.New("error")).Once()

	mes.handleNewMessage(testStream)

	assert.True(t, testStream.wasReset)
	assert.Len(t, buf.Bytes(), 0)
}

func TestP2PMessenger_handleNewMessage_NilHandlerResponse(t *testing.T) {
	mes, mocks := getMessengerWithMocks(t)

	requestEnvelope := &protocolpb.P2PEnvelope{
		Body: utils.RandomSlice(32),
	}

	requestEnvelopeBytes := getEncodedEnvelopeBytes(t, requestEnvelope)

	bufBytes := make([]byte, 0, 4096)
	buf := bytes.NewBuffer(bufBytes)

	connMock := p2pMocks.NewConnMock(t)

	peerID := libp2ppeer.ID("peer-id")

	connMock.On("RemotePeer").
		Return(peerID).Once()

	protocolID := protocol.ID("protocol-id")

	testStream := &testStream{
		r:          bytes.NewReader(requestEnvelopeBytes),
		w:          buf,
		conn:       connMock,
		protocolID: protocolID,
	}

	genericUtils.GetMock[*receiver.HandlerMock](mocks).
		On(
			"HandleInterceptor",
			mock.Anything,
			peerID,
			protocolID,
			mock.IsType(&protocolpb.P2PEnvelope{}),
		).Run(
		func(args mock.Arguments) {
			env := args.Get(3).(*protocolpb.P2PEnvelope)

			assert.Equal(t, requestEnvelope.GetBody(), env.GetBody())
		}).
		Return(nil, nil).Once()

	mes.handleNewMessage(testStream)

	assert.Len(t, buf.Bytes(), 0)
}

func TestP2PMessenger_handleNewMessage_WriteError(t *testing.T) {
	mes, mocks := getMessengerWithMocks(t)

	requestEnvelope := &protocolpb.P2PEnvelope{
		Body: utils.RandomSlice(32),
	}

	requestEnvelopeBytes := getEncodedEnvelopeBytes(t, requestEnvelope)

	bufBytes := make([]byte, 0, 0)
	buf := bytes.NewBuffer(bufBytes)

	connMock := p2pMocks.NewConnMock(t)

	peerID := libp2ppeer.ID("peer-id")

	connMock.On("RemotePeer").
		Return(peerID).Once()

	protocolID := protocol.ID("protocol-id")

	testStream := &testStream{
		r:          bytes.NewReader(requestEnvelopeBytes),
		w:          buf,
		conn:       connMock,
		protocolID: protocolID,
		failWrite:  true,
	}

	responseEnvelope := &protocolpb.P2PEnvelope{
		Body: utils.RandomSlice(32),
	}

	genericUtils.GetMock[*receiver.HandlerMock](mocks).
		On(
			"HandleInterceptor",
			mock.Anything,
			peerID,
			protocolID,
			mock.IsType(&protocolpb.P2PEnvelope{}),
		).Run(
		func(args mock.Arguments) {
			env := args.Get(3).(*protocolpb.P2PEnvelope)

			assert.Equal(t, requestEnvelope.GetBody(), env.GetBody())
		}).
		Return(responseEnvelope, nil).Once()

	mes.handleNewMessage(testStream)

	assert.True(t, testStream.wasReset)
	assert.Len(t, buf.Bytes(), 0)
}

func TestP2PMessenger_getMessageSender(t *testing.T) {
	mes, mocks := getMessengerWithMocks(t)

	peerID := libp2ppeer.ID("peer-id")
	protocolID := protocol.ID("protocol-id")

	senderMock := NewMessageSenderMock(t)

	args := &MessageSenderArgs{
		Ctx:        mes.ctx,
		Host:       mes.host,
		Timeout:    mes.timeout,
		PeerID:     peerID,
		ProtocolID: protocolID,
	}

	genericUtils.GetMock[*MessageSenderFactoryMock](mocks).
		On(
			"NewMessageSender",
			args,
		).Return(senderMock).Once()

	senderMock.On("Prepare").
		Return(nil).
		Once()

	res, err := mes.getMessageSender(peerID, protocolID)
	assert.NoError(t, err)
	assert.Equal(t, senderMock, res)
}

func TestP2PMessenger_getMessageSender_StoredSender(t *testing.T) {
	mes, _ := getMessengerWithMocks(t)

	peerID := libp2ppeer.ID("peer-id")
	protocolID := protocol.ID("protocol-id")

	senderMock := NewMessageSenderMock(t)

	senderMap := make(map[protocol.ID]MessageSender)
	senderMap[protocolID] = senderMock

	mes.strmap[peerID] = senderMap

	res, err := mes.getMessageSender(peerID, protocolID)
	assert.NoError(t, err)
	assert.Equal(t, senderMock, res)
}

func TestP2PMessenger_getMessageSender_StoredPeerID(t *testing.T) {
	mes, mocks := getMessengerWithMocks(t)

	peerID := libp2ppeer.ID("peer-id")
	protocolID := protocol.ID("protocol-id")

	mes.strmap[peerID] = make(map[protocol.ID]MessageSender)

	senderMock := NewMessageSenderMock(t)

	args := &MessageSenderArgs{
		Ctx:        mes.ctx,
		Host:       mes.host,
		Timeout:    mes.timeout,
		PeerID:     peerID,
		ProtocolID: protocolID,
	}

	genericUtils.GetMock[*MessageSenderFactoryMock](mocks).
		On(
			"NewMessageSender",
			args,
		).Return(senderMock).Once()

	senderMock.On("Prepare").
		Return(nil).
		Once()

	res, err := mes.getMessageSender(peerID, protocolID)
	assert.NoError(t, err)
	assert.Equal(t, senderMock, res)
}

func TestP2PMessenger_getMessageSender_StoredSenderAfterPrepareError(t *testing.T) {
	mes, mocks := getMessengerWithMocks(t)

	peerID := libp2ppeer.ID("peer-id")
	protocolID := protocol.ID("protocol-id")

	senderMock := NewMessageSenderMock(t)

	args := &MessageSenderArgs{
		Ctx:        mes.ctx,
		Host:       mes.host,
		Timeout:    mes.timeout,
		PeerID:     peerID,
		ProtocolID: protocolID,
	}

	genericUtils.GetMock[*MessageSenderFactoryMock](mocks).
		On(
			"NewMessageSender",
			args,
		).Return(senderMock).Once()

	senderMock2 := NewMessageSenderMock(t)

	senderMock.On("Prepare").
		Run(func(_ mock.Arguments) {
			mes.strmap[peerID] = make(map[protocol.ID]MessageSender)
			mes.strmap[peerID][protocolID] = senderMock2
		}).
		Return(errors.New("error")).
		Once()

	res, err := mes.getMessageSender(peerID, protocolID)
	assert.NoError(t, err)
	assert.Equal(t, senderMock2, res)
}

func TestP2PMessenger_getMessageSender_SameSenderAfterPrepareError(t *testing.T) {
	mes, mocks := getMessengerWithMocks(t)

	peerID := libp2ppeer.ID("peer-id")
	protocolID := protocol.ID("protocol-id")

	senderMock := NewMessageSenderMock(t)

	args := &MessageSenderArgs{
		Ctx:        mes.ctx,
		Host:       mes.host,
		Timeout:    mes.timeout,
		PeerID:     peerID,
		ProtocolID: protocolID,
	}

	genericUtils.GetMock[*MessageSenderFactoryMock](mocks).
		On(
			"NewMessageSender",
			args,
		).Return(senderMock).Once()

	senderMock.On("Prepare").
		Run(func(_ mock.Arguments) {
			mes.strmap[peerID] = make(map[protocol.ID]MessageSender)
			mes.strmap[peerID][protocolID] = senderMock
		}).
		Return(errors.New("error")).
		Once()

	res, err := mes.getMessageSender(peerID, protocolID)
	assert.NotNil(t, err)
	assert.Nil(t, res)
	assert.Nil(t, mes.strmap[peerID][protocolID])
}

func Test_writeAndReadMsg(t *testing.T) {
	b := make([]byte, 0, 4096)

	buf := bytes.NewBuffer(b)

	writer := bufio.NewWriter(buf)

	envelope := &protocolpb.P2PEnvelope{Body: utils.RandomSlice(32)}

	err := writeMsg(writer, envelope)
	assert.NoError(t, err)

	assertReaderContainsEnvelopeWithBody(t, bytes.NewReader(buf.Bytes()), envelope.GetBody())

	reader := bufio.NewReader(bytes.NewReader(buf.Bytes()))

	var res protocolpb.P2PEnvelope

	err = readMsg(reader, &res)
	assert.NoError(t, err)

	assert.Equal(t, envelope.GetBody(), res.GetBody())
}

func assertReaderContainsEnvelopeWithBody(t *testing.T, reader io.Reader, body []byte) {
	buf := bufio.NewReader(reader)

	msgLength, err := binary.ReadUvarint(buf)
	assert.NoError(t, err)

	msgBytes := make([]byte, msgLength)

	_, err = io.ReadFull(buf, msgBytes)

	var res protocolpb.P2PEnvelope

	err = proto.Unmarshal(msgBytes, &res)
	assert.NoError(t, err)

	assert.Equal(t, body, res.GetBody())
}

func getEncodedEnvelopeBytes(t *testing.T, envelope *protocolpb.P2PEnvelope) []byte {
	buf := make([]byte, MessageSizeMax)
	n := binary.PutUvarint(buf, uint64(proto.Size(envelope)))

	b, err := proto.Marshal(envelope)
	assert.NoError(t, err)

	buf = append(buf[:n], b...)

	return buf
}

func getMessengerWithMocks(t *testing.T) (*P2PMessenger, []any) {
	ctx := context.Background()
	hostMock := p2pMocks.NewHostMock(t)
	handlerMock := receiver.NewHandlerMock(t)
	factoryMock := NewMessageSenderFactoryMock(t)

	p2pMessenger := NewP2PMessenger(ctx, hostMock, time.Second, factoryMock, handlerMock.HandleInterceptor)

	return p2pMessenger.(*P2PMessenger), []any{
		hostMock,
		handlerMock,
		factoryMock,
	}
}

type testStream struct {
	inet.Stream

	r io.Reader
	w io.Writer

	conn       inet.Conn
	protocolID protocol.ID

	failWrite bool
	wasReset  bool
}

func (t *testStream) Conn() inet.Conn {
	return t.conn
}

func (t *testStream) Protocol() protocol.ID {
	return t.protocolID
}

func (t *testStream) Read(p []byte) (int, error) {
	return t.r.Read(p)
}

func (t *testStream) Write(p []byte) (int, error) {
	if t.failWrite {
		return 0, errors.New("write error")
	}

	return t.w.Write(p)
}

func (t *testStream) Reset() error {
	t.wasReset = true

	return nil
}
