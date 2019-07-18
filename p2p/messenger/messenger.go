package messenger

import (
	"bufio"
	"context"
	"io"
	"sync"
	"time"

	pb "github.com/centrifuge/centrifuge-protobufs/gen/go/protocol"
	"github.com/centrifuge/go-centrifuge/errors"
	ggio "github.com/gogo/protobuf/io"
	"github.com/golang/protobuf/proto"
	logging "github.com/ipfs/go-log"
	"github.com/jbenet/go-context/io"
	"github.com/libp2p/go-libp2p-host"
	inet "github.com/libp2p/go-libp2p-net"
	libp2pPeer "github.com/libp2p/go-libp2p-peer"
	"github.com/libp2p/go-libp2p-protocol"
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

type bufferedWriteCloser interface {
	ggio.WriteCloser
	Flush() error
}

// The Protobuf writer performs multiple small writes when writing a message.
// We need to buffer those writes, to make sure that we're not sending a new
// packet for every single write.
type bufferedDelimitedWriter struct {
	*bufio.Writer
	ggio.WriteCloser
}

func newBufferedDelimitedWriter(str io.Writer) bufferedWriteCloser {
	w := bufio.NewWriter(str)
	return &bufferedDelimitedWriter{
		Writer:      w,
		WriteCloser: ggio.NewDelimitedWriter(w),
	}
}

func (w *bufferedDelimitedWriter) Flush() error {
	return w.Writer.Flush()
}

// P2PMessenger is a libp2p messenger using protobufs and length delimited encoding
type P2PMessenger struct {
	host host.Host     // the network services we need
	self libp2pPeer.ID // Local peer (yourself)

	timeout time.Duration
	ctx     context.Context

	strmap map[libp2pPeer.ID]map[protocol.ID]*messageSender
	smlk   sync.Mutex

	plk sync.Mutex

	handler func(ctx context.Context, peer libp2pPeer.ID, protoc protocol.ID, msg *pb.P2PEnvelope) (*pb.P2PEnvelope, error)
}

// NewP2PMessenger returns a libp2p-messenger
func NewP2PMessenger(ctx context.Context, host host.Host, p2pTimeout time.Duration,
	handler func(ctx context.Context, peer libp2pPeer.ID, protoc protocol.ID, msg *pb.P2PEnvelope) (*pb.P2PEnvelope, error)) *P2PMessenger {
	return &P2PMessenger{
		ctx:     ctx,
		host:    host,
		self:    host.ID(),
		timeout: p2pTimeout,
		strmap:  make(map[libp2pPeer.ID]map[protocol.ID]*messageSender),
		handler: handler,
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
	// ok to use. we defer close stream in this func
	cr := ctxio.NewReader(ctx, s)
	cw := ctxio.NewWriter(ctx, s)

	// delimited readers and writers to set length of the protobuf messages to the stream
	r := ggio.NewDelimitedReader(cr, MessageSizeMax)
	w := newBufferedDelimitedWriter(cw)
	mPeer := s.Conn().RemotePeer()

	for {
		// receive msg
		pmes := new(pb.P2PEnvelope)
		switch err := r.ReadMsg(pmes); err {
		case io.EOF:
			s.Close()
			return
		case nil:
		default:
			s.Reset()
			log.Debugf("Error unmarshaling data: %s", err)
			return
		}

		if mes.handler == nil {
			s.Reset()
			log.Warning("got back nil handler from handlerForMsgType")
			return
		}

		// dispatch handler.
		rpmes, err := mes.handler(ctx, mPeer, s.Protocol(), pmes)
		if err != nil {
			s.Reset()
			log.Errorf("handle message error: %s", err)
			return
		}

		// if nil response, return it before serializing
		if rpmes == nil {
			log.Warning("got back nil response from request")
			continue
		}

		// send out response msg
		err = w.WriteMsg(rpmes)
		if err == nil {
			err = w.Flush()
		}
		if err != nil {
			s.Reset()
			log.Errorf("send response error: %s", err)
			return
		}
	}
}

// SendMessage sends out a request
func (mes *P2PMessenger) SendMessage(ctx context.Context, p libp2pPeer.ID, pmes *pb.P2PEnvelope, protoc protocol.ID) (*pb.P2PEnvelope, error) {
	ms, err := mes.messageSenderForPeerAndProto(p, protoc)
	if err != nil {
		return nil, err
	}

	rpmes, err := ms.sendMessage(ctx, pmes)
	if err != nil {
		return nil, err
	}

	return rpmes, nil
}

func (mes *P2PMessenger) messageSenderForPeerAndProto(p libp2pPeer.ID, protoc protocol.ID) (*messageSender, error) {
	mes.smlk.Lock()
	ms, ok := mes.strmap[p][protoc]
	if ok {
		mes.smlk.Unlock()
		return ms, nil
	}

	// create a new message sender for the peer and protocol
	ms = &messageSender{p: p, mes: mes, protoc: protoc}
	if mes.strmap[p] == nil {
		mes.strmap[p] = make(map[protocol.ID]*messageSender)
	}
	mes.strmap[p][protoc] = ms
	mes.smlk.Unlock()

	if err := ms.prepOrInvalidate(); err != nil {
		mes.smlk.Lock()
		defer mes.smlk.Unlock()

		if msCur, ok := mes.strmap[p][protoc]; ok {
			// Changed. Use the new one, old one is invalid and
			// not in the map so we can just throw it away.
			if ms != msCur {
				return msCur, nil
			}
			// Not changed, remove the now invalid stream from the
			// map.
			delete(mes.strmap[p], protoc)
		}
		// Invalid but not in map. Must have been removed by a disconnect.
		return nil, err
	}
	// All ready to go.
	return ms, nil
}

type messageSender struct {
	s      inet.Stream
	r      ggio.ReadCloser
	w      bufferedWriteCloser
	lk     sync.Mutex
	p      libp2pPeer.ID
	protoc protocol.ID
	mes    *P2PMessenger

	invalid   bool
	singleMes int
}

// invalidate is called before this messageSender is removed from the strmap.
// It prevents the messageSender from being reused/reinitialized and then
// forgotten (leaving the stream open).
func (ms *messageSender) invalidate() {
	ms.invalid = true
	if ms.s != nil {
		ms.s.Reset()
		ms.s = nil
	}
}

func (ms *messageSender) prepOrInvalidate() error {
	ms.lk.Lock()
	defer ms.lk.Unlock()
	if err := ms.prep(); err != nil {
		ms.invalidate()
		return err
	}
	return nil
}

func (ms *messageSender) prep() error {
	if ms.invalid {
		return ErrInvalidatedMessageSender
	}
	if ms.s != nil {
		return nil
	}

	// set the p2p timeout as the connection timeout
	timeoutCtx, canc := context.WithTimeout(ms.mes.ctx, ms.mes.timeout)
	defer canc()
	nstr, err := ms.mes.host.NewStream(timeoutCtx, ms.p, ms.protoc)
	if err != nil {
		return err
	}

	ms.r = ggio.NewDelimitedReader(nstr, MessageSizeMax)
	ms.w = newBufferedDelimitedWriter(nstr)
	ms.s = nstr
	return nil
}

// streamReuseTries is the number of times we will try to reuse a stream to a
// given peer before giving up and reverting to the old one-message-per-stream
// behaviour.
const streamReuseTries = 3

func (ms *messageSender) sendMessage(ctx context.Context, pmes *pb.P2PEnvelope) (*pb.P2PEnvelope, error) {
	ms.lk.Lock()
	defer ms.lk.Unlock()
	retry := false
	for {
		if err := ms.prep(); err != nil {
			return nil, err
		}

		if err := ms.writeMsg(pmes); err != nil {
			ms.s.Reset()
			ms.s = nil

			if retry {
				log.Info("error writing message, bailing: ", err)
				return nil, err
			}
			log.Info("error writing message, trying again: ", err)
			retry = true
			continue

		}

		mes := new(pb.P2PEnvelope)
		if err := ms.ctxReadMsg(ctx, mes); err != nil {
			ms.s.Reset()
			ms.s = nil

			if retry {
				log.Info("error reading message, bailing: ", err)
				return nil, err
			}
			log.Info("error reading message, trying again: ", err)
			retry = true
			continue

		}

		if ms.singleMes > streamReuseTries {
			go inet.FullClose(ms.s)
			ms.s = nil
		} else if retry {
			ms.singleMes++
		}

		return mes, nil
	}
}

func (ms *messageSender) writeMsg(pmes proto.Message) error {
	if err := ms.w.WriteMsg(pmes); err != nil {
		return err
	}
	return ms.w.Flush()
}

func (ms *messageSender) ctxReadMsg(ctx context.Context, mes *pb.P2PEnvelope) error {
	errc := make(chan error, 1)
	go func(r ggio.ReadCloser) {
		errc <- r.ReadMsg(mes)
	}(ms.r)

	t := time.NewTimer(ms.mes.timeout)
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
