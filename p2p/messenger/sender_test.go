//go:build unit

package messenger

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"math/rand"
	"testing"
	"time"

	protocolpb "github.com/centrifuge/centrifuge-protobufs/gen/go/protocol"
	p2pMocks "github.com/centrifuge/go-centrifuge/p2p/mocks"
	genericUtils "github.com/centrifuge/go-centrifuge/testingutils/generic"
	"github.com/centrifuge/go-centrifuge/utils"
	libp2ppeer "github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/protocol"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestMessageSender_Prepare(t *testing.T) {
	ms, mocks := getMessageSenderWithMocks(t)
	ms.stream = nil

	expectedCtx, cancel := context.WithTimeout(ms.ctx, ms.timeout)
	defer cancel()

	streamMock := p2pMocks.NewStreamMock(t)

	genericUtils.GetMock[*p2pMocks.HostMock](mocks).
		On(
			"NewStream",
			mock.IsType(expectedCtx),
			ms.peerID,
			ms.protocolID,
		).Return(streamMock, nil).Once()

	err := ms.Prepare()
	assert.NoError(t, err)
}

func TestMessageSender_Prepare_InvalidMessageSender(t *testing.T) {
	ms, _ := getMessageSenderWithMocks(t)
	ms.stream = nil
	ms.invalid = true

	err := ms.Prepare()
	assert.ErrorIs(t, err, ErrInvalidatedMessageSender)
}

func TestMessageSender_Prepare_InvalidMessageSender_WithStream(t *testing.T) {
	ms, mocks := getMessageSenderWithMocks(t)
	ms.invalid = true

	genericUtils.GetMock[*p2pMocks.StreamMock](mocks).
		On("Reset").
		Return(nil).
		Once()

	err := ms.Prepare()
	assert.ErrorIs(t, err, ErrInvalidatedMessageSender)
}

func TestMessageSender_Prepare_StreamPresent(t *testing.T) {
	ms, _ := getMessageSenderWithMocks(t)

	err := ms.Prepare()
	assert.NoError(t, err)
}

func TestMessageSender_Prepare_HostError(t *testing.T) {
	ms, mocks := getMessageSenderWithMocks(t)
	ms.stream = nil

	expectedCtx, cancel := context.WithTimeout(ms.ctx, ms.timeout)
	defer cancel()

	hostErr := errors.New("error")
	genericUtils.GetMock[*p2pMocks.HostMock](mocks).
		On(
			"NewStream",
			mock.IsType(expectedCtx),
			ms.peerID,
			ms.protocolID,
		).Return(nil, hostErr).Once()

	err := ms.Prepare()
	assert.ErrorIs(t, err, hostErr)
	assert.True(t, ms.invalid)
}

func TestMessageSender_SendMessage(t *testing.T) {
	ms, _ := getMessageSenderWithMocks(t)

	var b []byte

	buf := bytes.NewBuffer(b)

	ms.writer = bufio.NewWriter(buf)

	incomingEnvelope := &protocolpb.P2PEnvelope{
		Body: utils.RandomSlice(32),
	}

	incomingEnvelopeBytes := getEncodedEnvelopeBytes(t, incomingEnvelope)

	ms.reader = bufio.NewReader(bytes.NewReader(incomingEnvelopeBytes))

	outgoingEnvelope := &protocolpb.P2PEnvelope{
		Body: utils.RandomSlice(32),
	}

	ctx := context.Background()

	res, err := ms.SendMessage(ctx, outgoingEnvelope)
	assert.NoError(t, err)
	assert.Equal(t, incomingEnvelope.GetBody(), res.GetBody())

	assertReaderContainsEnvelopeWithBody(t, bytes.NewReader(buf.Bytes()), outgoingEnvelope.GetBody())
}

func TestMessageSender_SendMessage_PrepError(t *testing.T) {
	ms, _ := getMessageSenderWithMocks(t)
	ms.invalid = true

	outgoingEnvelope := &protocolpb.P2PEnvelope{
		Body: utils.RandomSlice(32),
	}

	ctx := context.Background()

	res, err := ms.SendMessage(ctx, outgoingEnvelope)
	assert.NotNil(t, err)
	assert.Nil(t, res)
}

func TestMessageSender_SendMessage_WriteError(t *testing.T) {
	ms, mocks := getMessageSenderWithMocks(t)

	ms.writer = bufio.NewWriter(&errorReadWriter{})

	outgoingEnvelope := &protocolpb.P2PEnvelope{
		Body: utils.RandomSlice(32),
	}

	ctx := context.Background()

	genericUtils.GetMock[*p2pMocks.StreamMock](mocks).
		On("Reset").
		Return(nil).
		Once()

	newStreamMock := p2pMocks.NewStreamMock(t)

	expectedCtx, cancel := context.WithTimeout(ms.ctx, ms.timeout)
	defer cancel()

	genericUtils.GetMock[*p2pMocks.HostMock](mocks).
		On(
			"NewStream",
			mock.IsType(expectedCtx),
			ms.peerID,
			ms.protocolID,
		).Return(newStreamMock, nil).Once()

	streamWriteErr := errors.New("error")

	newStreamMock.On("Write", mock.IsType([]byte{})).
		Return(0, streamWriteErr).Once()

	newStreamMock.
		On("Reset").
		Return(nil).
		Once()

	res, err := ms.SendMessage(ctx, outgoingEnvelope)
	assert.ErrorIs(t, err, streamWriteErr)
	assert.Nil(t, res)
	assert.Nil(t, ms.stream)
}

func TestMessageSender_SendMessage_WriteError_NewStreamError(t *testing.T) {
	ms, mocks := getMessageSenderWithMocks(t)

	ms.writer = bufio.NewWriter(&errorReadWriter{})

	outgoingEnvelope := &protocolpb.P2PEnvelope{
		Body: utils.RandomSlice(32),
	}

	ctx := context.Background()

	genericUtils.GetMock[*p2pMocks.StreamMock](mocks).
		On("Reset").
		Return(nil).
		Once()

	newStreamErr := errors.New("error")

	expectedCtx, cancel := context.WithTimeout(ms.ctx, ms.timeout)
	defer cancel()

	genericUtils.GetMock[*p2pMocks.HostMock](mocks).
		On(
			"NewStream",
			mock.IsType(expectedCtx),
			ms.peerID,
			ms.protocolID,
		).Return(nil, newStreamErr).Once()

	res, err := ms.SendMessage(ctx, outgoingEnvelope)
	assert.ErrorIs(t, err, newStreamErr)
	assert.Nil(t, res)
	assert.Nil(t, ms.stream)
}

func TestMessageSender_SendMessage_ReadError(t *testing.T) {
	ms, mocks := getMessageSenderWithMocks(t)

	var b []byte

	buf := bytes.NewBuffer(b)

	ms.writer = bufio.NewWriter(buf)

	ms.reader = bufio.NewReader(&errorReadWriter{})

	outgoingEnvelope := &protocolpb.P2PEnvelope{
		Body: utils.RandomSlice(32),
	}

	ctx := context.Background()

	genericUtils.GetMock[*p2pMocks.StreamMock](mocks).
		On("Reset").
		Return(nil).
		Once()

	newStreamMock := p2pMocks.NewStreamMock(t)

	expectedCtx, cancel := context.WithTimeout(ms.ctx, ms.timeout)
	defer cancel()

	genericUtils.GetMock[*p2pMocks.HostMock](mocks).
		On(
			"NewStream",
			mock.IsType(expectedCtx),
			ms.peerID,
			ms.protocolID,
		).Return(newStreamMock, nil).Once()

	newStreamMock.On("Write", mock.IsType([]byte{})).
		Return(rand.Int(), nil).Once()

	streamReadErr := errors.New("error")

	newStreamMock.On("Read", mock.IsType([]byte{})).
		Return(0, streamReadErr).Once()

	newStreamMock.
		On("Reset").
		Return(nil).
		Once()

	res, err := ms.SendMessage(ctx, outgoingEnvelope)
	assert.ErrorIs(t, err, streamReadErr)
	assert.Nil(t, res)
	assert.Nil(t, ms.stream)
}

func TestMessageSender_SendMessage_ReadError_NewStreamError(t *testing.T) {
	ms, mocks := getMessageSenderWithMocks(t)

	var b []byte

	buf := bytes.NewBuffer(b)

	ms.writer = bufio.NewWriter(buf)

	ms.reader = bufio.NewReader(&errorReadWriter{})

	outgoingEnvelope := &protocolpb.P2PEnvelope{
		Body: utils.RandomSlice(32),
	}

	ctx := context.Background()

	genericUtils.GetMock[*p2pMocks.StreamMock](mocks).
		On("Reset").
		Return(nil).
		Once()

	expectedCtx, cancel := context.WithTimeout(ms.ctx, ms.timeout)
	defer cancel()

	newStreamError := errors.New("error")

	genericUtils.GetMock[*p2pMocks.HostMock](mocks).
		On(
			"NewStream",
			mock.IsType(expectedCtx),
			ms.peerID,
			ms.protocolID,
		).Return(nil, newStreamError).Once()

	res, err := ms.SendMessage(ctx, outgoingEnvelope)
	assert.ErrorIs(t, err, newStreamError)
	assert.Nil(t, res)
	assert.Nil(t, ms.stream)
}

func TestMessageSender_SendMessage_TooManyUses(t *testing.T) {
	ms, mocks := getMessageSenderWithMocks(t)
	ms.currentStreamUses = maxStreamReuseTries

	ms.writer = bufio.NewWriter(&errorReadWriter{})

	incomingEnvelope := &protocolpb.P2PEnvelope{
		Body: utils.RandomSlice(32),
	}

	incomingEnvelopeBytes := getEncodedEnvelopeBytes(t, incomingEnvelope)

	ms.reader = bufio.NewReader(bytes.NewReader(incomingEnvelopeBytes))

	outgoingEnvelope := &protocolpb.P2PEnvelope{
		Body: utils.RandomSlice(32),
	}

	ctx := context.Background()

	genericUtils.GetMock[*p2pMocks.StreamMock](mocks).
		On("Reset").
		Return(nil).
		Once()

	newStreamMock := p2pMocks.NewStreamMock(t)

	expectedCtx, cancel := context.WithTimeout(ms.ctx, ms.timeout)
	defer cancel()

	genericUtils.GetMock[*p2pMocks.HostMock](mocks).
		On(
			"NewStream",
			mock.IsType(expectedCtx),
			ms.peerID,
			ms.protocolID,
		).Return(newStreamMock, nil).Once()

	newStreamMock.On("Write", mock.IsType([]byte{})).
		Return(rand.Int(), nil).Once()

	newStreamMock.On("Read", mock.IsType([]byte{})).
		Return(rand.Int(), nil).Once()

	res, err := ms.SendMessage(ctx, outgoingEnvelope)
	assert.NoError(t, err)
	assert.NotNil(t, res)
}

func TestMessageSender_ctxReadMsg_Error(t *testing.T) {
	ms, _ := getMessageSenderWithMocks(t)

	// Context error

	ms.reader = bufio.NewReader(&errorReadWriter{readTimeout: 5 * time.Second})

	ms.timeout = 3 * time.Second

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)

	msg := &protocolpb.P2PEnvelope{}

	err := ms.ctxReadMsg(ctx, msg)
	assert.NotNil(t, err)

	cancel()

	// Read timeout error

	ms, _ = getMessageSenderWithMocks(t)

	ms.reader = bufio.NewReader(&errorReadWriter{readTimeout: 5 * time.Second})

	ms.timeout = 1 * time.Second

	ctx, cancel = context.WithTimeout(context.Background(), 3*time.Second)

	err = ms.ctxReadMsg(ctx, msg)
	assert.ErrorIs(t, err, ErrReadTimeout)

	cancel()

	// Reader error

	ms, _ = getMessageSenderWithMocks(t)

	ms.reader = bufio.NewReader(&errorReadWriter{readTimeout: 5 * time.Second})

	ms.reader = bufio.NewReader(&errorReadWriter{readTimeout: 0 * time.Second})
	ms.timeout = 3 * time.Second
	ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)

	err = ms.ctxReadMsg(ctx, msg)
	assert.ErrorIs(t, err, readWriteError)

	cancel()
}

var (
	readWriteError = errors.New("error")
)

type errorReadWriter struct {
	readTimeout time.Duration
}

func (rw *errorReadWriter) Write(_ []byte) (int, error) {
	return 0, readWriteError
}

func (rw *errorReadWriter) Read(_ []byte) (int, error) {
	time.Sleep(rw.readTimeout)

	return 0, readWriteError
}

func getMessageSenderWithMocks(t *testing.T) (*messageSender, []any) {
	ctx := context.Background()
	hostMock := p2pMocks.NewHostMock(t)
	streamMock := p2pMocks.NewStreamMock(t)
	timeout := 1 * time.Second
	peerID := libp2ppeer.ID("peer-id")
	protocolID := protocol.ID("protocol-id")

	ms := &messageSender{
		ctx:        ctx,
		stream:     streamMock,
		peerID:     peerID,
		protocolID: protocolID,
		timeout:    timeout,
		host:       hostMock,
	}

	return ms, []any{
		hostMock,
		streamMock,
	}
}
