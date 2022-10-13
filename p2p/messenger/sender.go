package messenger

import (
	"bufio"
	"context"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p-core/host"

	pb "github.com/centrifuge/centrifuge-protobufs/gen/go/protocol"
	inet "github.com/libp2p/go-libp2p-core/network"
	libp2pPeer "github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/protocol"
)

type MessageSenderArgs struct {
	Ctx        context.Context
	Host       host.Host
	Timeout    time.Duration
	PeerID     libp2pPeer.ID
	ProtocolID protocol.ID
}

//go:generate mockery --name MessageSenderFactory --structname MessageSenderFactoryMock --filename factory_mock.go --inpackage

type MessageSenderFactory interface {
	NewMessageSender(args *MessageSenderArgs) MessageSender
}

type messageSenderFactory struct{}

func (f *messageSenderFactory) NewMessageSender(args *MessageSenderArgs) MessageSender {
	return &messageSender{
		ctx:        args.Ctx,
		host:       args.Host,
		timeout:    args.Timeout,
		peerID:     args.PeerID,
		protocolID: args.ProtocolID,
	}
}

func NewMessageSenderFactory() MessageSenderFactory {
	return &messageSenderFactory{}
}

//go:generate mockery --name MessageSender --structname MessageSenderMock --filename sender_mock.go --inpackage

type MessageSender interface {
	Prepare() error
	SendMessage(ctx context.Context, pmes *pb.P2PEnvelope) (*pb.P2PEnvelope, error)
}
type messageSender struct {
	ctx        context.Context
	stream     inet.Stream
	reader     *bufio.Reader
	writer     *bufio.Writer
	mutex      sync.Mutex
	peerID     libp2pPeer.ID
	protocolID protocol.ID
	timeout    time.Duration
	host       host.Host

	invalid           bool
	currentStreamUses int
}

func (ms *messageSender) Prepare() error {
	ms.mutex.Lock()
	defer ms.mutex.Unlock()
	if err := ms.prep(); err != nil {
		ms.invalidate()
		return err
	}
	return nil
}

// maxStreamReuseTries is the number of times we will try to reuse a stream to a
// given peer before giving up and reverting to the old one-message-per-stream
// behaviour.
const maxStreamReuseTries = 3

func (ms *messageSender) SendMessage(ctx context.Context, pmes *pb.P2PEnvelope) (*pb.P2PEnvelope, error) {
	ms.mutex.Lock()
	defer ms.mutex.Unlock()

	disableRetry := false

	for {
		if err := ms.prep(); err != nil {
			return nil, err
		}

		if err := writeMsg(ms.writer, pmes); err != nil {
			ms.stream.Reset()
			ms.stream = nil

			if disableRetry {
				log.Info("error writing message, bailing: ", err)
				return nil, err
			}
			log.Info("error writing message, trying again: ", err)
			disableRetry = true
			continue
		}

		mes := new(pb.P2PEnvelope)
		if err := ms.ctxReadMsg(ctx, mes); err != nil {
			ms.stream.Reset()
			ms.stream = nil

			if disableRetry {
				log.Info("error reading message, bailing: ", err)
				return nil, err
			}
			log.Info("error reading message, trying again: ", err)
			disableRetry = true
			continue
		}

		if ms.currentStreamUses > maxStreamReuseTries {
			log.Infof("too many uses, closing stream: %s", ms.stream.Close())
			ms.stream = nil
		} else if disableRetry {
			ms.currentStreamUses++
		}

		return mes, nil
	}
}

// invalidate is called before this messageSender is removed from the strmap.
// It prevents the messageSender from being reused/reinitialized and then
// forgotten (leaving the stream open).
func (ms *messageSender) invalidate() {
	ms.invalid = true
	if ms.stream != nil {
		ms.stream.Reset()
		ms.stream = nil
	}
}

func (ms *messageSender) prep() error {
	if ms.invalid {
		return ErrInvalidatedMessageSender
	}
	if ms.stream != nil {
		return nil
	}

	// set the p2p timeout as the connection timeout
	timeoutCtx, canc := context.WithTimeout(ms.ctx, ms.timeout)
	defer canc()

	nstr, err := ms.host.NewStream(timeoutCtx, ms.peerID, ms.protocolID)
	if err != nil {
		return err
	}

	ms.reader = bufio.NewReader(nstr)
	ms.writer = bufio.NewWriter(nstr)
	ms.stream = nstr
	return nil
}

func (ms *messageSender) ctxReadMsg(ctx context.Context, mes *pb.P2PEnvelope) error {
	errc := make(chan error, 1)
	go func(r *bufio.Reader) {
		errc <- readMsg(r, mes)
	}(ms.reader)

	t := time.NewTimer(ms.timeout)
	defer t.Stop()

	select {
	case err := <-errc:
		return err
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return ErrReadTimeout
	}
}
