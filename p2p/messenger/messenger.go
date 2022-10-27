package messenger

import (
	"bufio"
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"sync"
	"time"

	pb "github.com/centrifuge/centrifuge-protobufs/gen/go/protocol"
	"github.com/centrifuge/go-centrifuge/errors"
	logging "github.com/ipfs/go-log"
	"github.com/libp2p/go-libp2p-core/host"
	inet "github.com/libp2p/go-libp2p-core/network"
	libp2pPeer "github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/protocol"
	"google.golang.org/protobuf/proto"
)

const (
	// MessageSizeMax is a soft maximum for network messages.
	MessageSizeMax = 1 << 25 // 32 MB

	// ErrReadTimeout must be used when receiving timeout while reading
	ErrReadTimeout = errors.Error("timed out reading response")

	// ErrInvalidatedMessageSender must be used when the message sender object created is no longer valid (connection has dropped)
	ErrInvalidatedMessageSender = errors.Error("message sender has been invalidated")
)

var log = logging.Logger("p2p-messenger")

//go:generate mockery --name Messenger --structname MessengerMock --filename messenger_mock.go --inpackage

type Messenger interface {
	Init(protocols ...protocol.ID)
	SendMessage(ctx context.Context, peerID libp2pPeer.ID, mes *pb.P2PEnvelope, protocolID protocol.ID) (*pb.P2PEnvelope, error)
}

// P2PMessenger is a libp2p messenger using protobufs and length delimited encoding
type P2PMessenger struct {
	host host.Host

	timeout time.Duration
	ctx     context.Context

	strmap map[libp2pPeer.ID]map[protocol.ID]MessageSender
	smlk   sync.Mutex

	messageSenderFactory MessageSenderFactory

	handler func(ctx context.Context, peer libp2pPeer.ID, protoc protocol.ID, msg *pb.P2PEnvelope) (*pb.P2PEnvelope, error)
}

// NewP2PMessenger returns a libp2p-messenger
func NewP2PMessenger(
	ctx context.Context,
	host host.Host,
	p2pTimeout time.Duration,
	messageSenderFactory MessageSenderFactory,
	handler func(ctx context.Context, peer libp2pPeer.ID, protoc protocol.ID, msg *pb.P2PEnvelope) (*pb.P2PEnvelope, error),
) Messenger {
	return &P2PMessenger{
		ctx:                  ctx,
		host:                 host,
		timeout:              p2pTimeout,
		strmap:               make(map[libp2pPeer.ID]map[protocol.ID]MessageSender),
		messageSenderFactory: messageSenderFactory,
		handler:              handler,
	}
}

// Init initiates listening to given set of protocol streams
func (mes *P2PMessenger) Init(protocols ...protocol.ID) {
	for _, p := range protocols {
		mes.host.SetStreamHandler(p, mes.handleNewStream)
	}
}

// handleNewStream implements the inet.StreamHandler
func (mes *P2PMessenger) handleNewStream(s inet.Stream) {
	go mes.handleNewMessage(s)
}

func (mes *P2PMessenger) handleNewMessage(s inet.Stream) {
	ctx := mes.ctx
	w := bufio.NewWriter(s)
	mPeer := s.Conn().RemotePeer()

	for {
		r := bufio.NewReader(s)

		var pmes pb.P2PEnvelope

		if err := readMsg(r, &pmes); err != nil {
			log.Errorf("Couldn't read message: %s", err)
			s.Reset()
			return
		}

		if mes.handler == nil {
			s.Reset()
			log.Warn("No message handler")
			return
		}

		rpmes, err := mes.handler(ctx, mPeer, s.Protocol(), &pmes)
		if err != nil {
			s.Reset()
			log.Errorf("Couldn't handle message: %s", err)
			return
		}

		if rpmes == nil {
			log.Warn("No response from handler")
			continue
		}

		if err := writeMsg(w, rpmes); err != nil {
			log.Errorf("Couldn't write response message: %s", err)
			s.Reset()
			return
		}
	}
}

// SendMessage sends out a request
func (mes *P2PMessenger) SendMessage(ctx context.Context, peerID libp2pPeer.ID, pmes *pb.P2PEnvelope, protocolID protocol.ID) (*pb.P2PEnvelope, error) {
	ms, err := mes.getMessageSender(peerID, protocolID)
	if err != nil {
		log.Errorf("Couldn't get message sender: %s", err)

		return nil, err
	}

	rpmes, err := ms.SendMessage(ctx, pmes)
	if err != nil {
		log.Errorf("Couldn't get message sender: %s", err)

		return nil, err
	}

	return rpmes, nil
}

func (mes *P2PMessenger) getMessageSender(peerID libp2pPeer.ID, protocolID protocol.ID) (MessageSender, error) {
	mes.smlk.Lock()
	ms, ok := mes.strmap[peerID][protocolID]
	if ok {
		mes.smlk.Unlock()
		return ms, nil
	}

	args := &MessageSenderArgs{
		Ctx:        mes.ctx,
		Host:       mes.host,
		Timeout:    mes.timeout,
		PeerID:     peerID,
		ProtocolID: protocolID,
	}

	// create a new message sender for the peer and protocol
	ms = mes.messageSenderFactory.NewMessageSender(args)
	if mes.strmap[peerID] == nil {
		mes.strmap[peerID] = make(map[protocol.ID]MessageSender)
	}
	mes.strmap[peerID][protocolID] = ms
	mes.smlk.Unlock()

	if err := ms.Prepare(); err != nil {
		mes.smlk.Lock()
		defer mes.smlk.Unlock()

		log.Errorf("Couldn't prepare message sender: %s", err)

		if msCur, ok := mes.strmap[peerID][protocolID]; ok {
			// Changed. Use the new one, old one is invalid and
			// not in the map so we can just throw it away.
			if ms != msCur {
				return msCur, nil
			}
			// Not changed, remove the now invalid stream from the
			// map.
			delete(mes.strmap[peerID], protocolID)
		}
		// Invalid but not in map. Must have been removed by a disconnect.
		return nil, err
	}
	// All ready to go.
	return ms, nil
}

func writeMsg(w *bufio.Writer, msg proto.Message) error {
	buf := make([]byte, MessageSizeMax)

	n := binary.PutUvarint(buf, uint64(proto.Size(msg)))

	b, err := proto.Marshal(msg)

	if err != nil {
		return fmt.Errorf("couldn't marshal message: %w", err)
	}

	buf = append(buf[:n], b...)

	if _, err := w.Write(buf); err != nil {
		return fmt.Errorf("couldn't write to buffer: %w", err)
	}

	if err := w.Flush(); err != nil {
		return fmt.Errorf("couldn't flush writer: %w", err)
	}

	return nil
}

func readMsg(r *bufio.Reader, msg proto.Message) error {
	length, err := binary.ReadUvarint(r)

	if err != nil {
		return fmt.Errorf("couldn't read message length: %w", err)
	}

	if length > MessageSizeMax {
		return fmt.Errorf("message too big - %d", length)
	}

	b := make([]byte, length)

	if _, err := io.ReadFull(r, b); err != nil {
		return fmt.Errorf("couldn't read message: %w", err)
	}

	if err := proto.Unmarshal(b, msg); err != nil {
		return fmt.Errorf("couldn't unmarshal message: %w", err)
	}

	return nil
}
