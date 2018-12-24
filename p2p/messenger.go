package p2p

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/centrifuge/go-centrifuge/errors"

	"github.com/libp2p/go-libp2p-host"
	"github.com/libp2p/go-libp2p-protocol"

	pb "github.com/centrifuge/go-centrifuge/protobufs/gen/go/protocol"
	ggio "github.com/gogo/protobuf/io"
	"github.com/jbenet/go-context/io"
	inet "github.com/libp2p/go-libp2p-net"
	peer "github.com/libp2p/go-libp2p-peer"
	ps "github.com/libp2p/go-libp2p-peerstore"
)

// ErrReadTimeout timeout while reading
const ErrReadTimeout = errors.Error("timed out reading response")

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

type p2pMessenger struct {
	host host.Host // the network services we need
	self peer.ID   // Local peer (yourself)

	timout time.Duration
	ctx    context.Context

	strmap map[peer.ID]*messageSender
	smlk   sync.Mutex

	plk sync.Mutex

	msgHandlers map[pb.MessageType]func(ctx context.Context, peer peer.ID, protoc protocol.ID, msg *pb.P2PEnvelope) (*pb.P2PEnvelope, error)
}

func newMessenger(ctx context.Context, host host.Host, self peer.ID, p2pTimeout time.Duration) *p2pMessenger {
	return &p2pMessenger{
		ctx:         ctx,
		host:        host,
		self:        self,
		timout:      p2pTimeout,
		strmap:      make(map[peer.ID]*messageSender),
		msgHandlers: make(map[pb.MessageType]func(ctx context.Context, peer peer.ID, protoc protocol.ID, msg *pb.P2PEnvelope) (*pb.P2PEnvelope, error))}
}

// addHandler adds a message handler for a specific message type
func (mes *p2pMessenger) addHandler(mType pb.MessageType, handler func(ctx context.Context, peer peer.ID, protoc protocol.ID, msg *pb.P2PEnvelope) (*pb.P2PEnvelope, error)) {
	mes.msgHandlers[mType] = handler
}

// handleNewStream implements the inet.StreamHandler
func (mes *p2pMessenger) handleNewStream(s inet.Stream) {
	go mes.handleNewMessage(s)
}

func (mes *p2pMessenger) handleNewMessage(s inet.Stream) {
	ctx := mes.ctx
	cr := ctxio.NewReader(ctx, s) // ok to use. we defer close stream in this func
	cw := ctxio.NewWriter(ctx, s) // ok to use. we defer close stream in this func
	r := ggio.NewDelimitedReader(cr, inet.MessageSizeMax)
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

		// get handler for this msg type.
		handler := mes.msgHandlers[pmes.GetType()]
		if handler == nil {
			s.Reset()
			log.Debug("got back nil handler from handlerForMsgType")
			return
		}

		// dispatch handler.
		rpmes, err := handler(ctx, mPeer, s.Protocol(), pmes)
		if err != nil {
			s.Reset()
			log.Debugf("handle message error: %s", err)
			return
		}

		// if nil response, return it before serializing
		if rpmes == nil {
			log.Debug("got back nil response from request")
			continue
		}

		// send out response msg
		err = w.WriteMsg(rpmes)
		if err == nil {
			err = w.Flush()
		}
		if err != nil {
			s.Reset()
			log.Debugf("send response error: %s", err)
			return
		}
	}
}

// sendRequest sends out a request
func (mes *p2pMessenger) sendRequest(ctx context.Context, p peer.ID, pmes *pb.P2PEnvelope, protoc protocol.ID) (*pb.P2PEnvelope, error) {

	ms, err := mes.messageSenderForPeer(p, protoc)
	if err != nil {
		return nil, err
	}

	rpmes, err := ms.SendRequest(ctx, pmes, protoc)
	if err != nil {
		return nil, err
	}

	return rpmes, nil
}

// sendMessage sends out a message
func (mes *p2pMessenger) sendMessage(ctx context.Context, p peer.ID, pmes *pb.P2PEnvelope, protoc protocol.ID) error {
	ms, err := mes.messageSenderForPeer(p, protoc)
	if err != nil {
		return err
	}

	if err := ms.SendMessage(ctx, pmes, protoc); err != nil {
		return err
	}
	return nil
}

func (mes *p2pMessenger) messageSenderForPeer(p peer.ID, protoc protocol.ID) (*messageSender, error) {
	mes.smlk.Lock()
	ms, ok := mes.strmap[p]
	if ok {
		mes.smlk.Unlock()
		return ms, nil
	}
	ms = &messageSender{p: p, mes: mes}
	mes.strmap[p] = ms
	mes.smlk.Unlock()

	if err := ms.prepOrInvalidate(protoc); err != nil {
		mes.smlk.Lock()
		defer mes.smlk.Unlock()

		if msCur, ok := mes.strmap[p]; ok {
			// Changed. Use the new one, old one is invalid and
			// not in the map so we can just throw it away.
			if ms != msCur {
				return msCur, nil
			}
			// Not changed, remove the now invalid stream from the
			// map.
			delete(mes.strmap, p)
		}
		// Invalid but not in map. Must have been removed by a disconnect.
		return nil, err
	}
	// All ready to go.
	return ms, nil
}

type messageSender struct {
	s   inet.Stream
	r   ggio.ReadCloser
	w   bufferedWriteCloser
	lk  sync.Mutex
	p   peer.ID
	mes *p2pMessenger

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

func (ms *messageSender) prepOrInvalidate(protoc protocol.ID) error {
	ms.lk.Lock()
	defer ms.lk.Unlock()
	if err := ms.prep(protoc); err != nil {
		ms.invalidate()
		return err
	}
	return nil
}

func (ms *messageSender) prep(protoc protocol.ID) error {
	if ms.invalid {
		return fmt.Errorf("message sender has been invalidated")
	}
	if ms.s != nil {
		return nil
	}

	err := ms.mes.host.Connect(ms.mes.ctx, ps.PeerInfo{
		ID: ms.p,
	})
	if err != nil {
		return err
	}

	nstr, err := ms.mes.host.NewStream(ms.mes.ctx, ms.p, protoc)
	if err != nil {
		return err
	}

	ms.r = ggio.NewDelimitedReader(nstr, inet.MessageSizeMax)
	ms.w = newBufferedDelimitedWriter(nstr)
	ms.s = nstr

	return nil
}

// streamReuseTries is the number of times we will try to reuse a stream to a
// given peer before giving up and reverting to the old one-message-per-stream
// behaviour.
const streamReuseTries = 3

func (ms *messageSender) SendMessage(ctx context.Context, pmes *pb.P2PEnvelope, protoc protocol.ID) error {
	ms.lk.Lock()
	defer ms.lk.Unlock()
	retry := false
	for {
		if err := ms.prep(protoc); err != nil {
			return err
		}

		if err := ms.writeMsg(pmes); err != nil {
			ms.s.Reset()
			ms.s = nil

			if retry {
				log.Info("error writing message, bailing: ", err)
				return err
			}
			log.Info("error writing message, trying again: ", err)
			retry = true
			continue

		}

		if ms.singleMes > streamReuseTries {
			go inet.FullClose(ms.s)
			ms.s = nil
		} else if retry {
			ms.singleMes++
		}

		return nil
	}
}

func (ms *messageSender) SendRequest(ctx context.Context, pmes *pb.P2PEnvelope, protoc protocol.ID) (*pb.P2PEnvelope, error) {
	ms.lk.Lock()
	defer ms.lk.Unlock()
	retry := false
	for {
		if err := ms.prep(protoc); err != nil {
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

func (ms *messageSender) writeMsg(pmes *pb.P2PEnvelope) error {
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

	t := time.NewTimer(ms.mes.timout)
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
