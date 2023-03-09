package dispatcher

import (
	"context"
	"time"

	"github.com/centrifuge/pod/errors"
	logging "github.com/ipfs/go-log"
)

var (
	log = logging.Logger("dispatcher")
)

const (
	ErrDispatchContextDone   = errors.Error("dispatch context is done")
	ErrDispatcherContextDone = errors.Error("dispatcher context is done")
	ErrSubscribeContextDone  = errors.Error("subscribe context is done")
)

//go:generate mockery --name Dispatcher --structname DispatcherMock --filename dispatcher_mock.go --inpackage

type Dispatcher[T any] interface {
	Dispatch(ctx context.Context, t T) error

	Subscribe(context.Context) (chan T, error)
	Unsubscribe(chan T) error

	Stop()
}

type dispatcher[T any] struct {
	ctx      context.Context
	cancelFn context.CancelFunc

	dispatch        chan T
	subscribeChan   chan chan T
	unsubscribeChan chan chan T

	opts *opts
}

func NewDispatcher[T any](ctx context.Context, opts ...Opt) Dispatcher[T] {
	dispatcherOpts := getDefaultOpts()

	for _, opt := range opts {
		opt(dispatcherOpts)
	}

	dispatchChan := make(chan T, dispatcherOpts.dispatchChanBuf)
	subscribeChan := make(chan chan T, dispatcherOpts.subscribeChanBuf)
	unsubscribeChan := make(chan chan T, dispatcherOpts.unsubscribeChanBuf)

	ctx, canc := context.WithCancel(ctx)

	d := &dispatcher[T]{
		ctx,
		canc,
		dispatchChan,
		subscribeChan,
		unsubscribeChan,
		dispatcherOpts,
	}

	go d.run()

	return d
}

func (d *dispatcher[T]) Subscribe(ctx context.Context) (chan T, error) {
	c := make(chan T, d.opts.subscriberChanBuf)

	select {
	case <-d.ctx.Done():
		log.Errorf("Dispatcher context is done: %s", d.ctx.Err())
		return nil, ErrDispatcherContextDone
	case <-ctx.Done():
		log.Errorf("Subscribe context is done: %s", ctx.Err())
		return nil, ErrSubscribeContextDone
	case d.subscribeChan <- c:
		return c, nil
	}
}

func (d *dispatcher[T]) Unsubscribe(c chan T) error {
	select {
	case <-d.ctx.Done():
		log.Errorf("Dispatcher context is done: %s", d.ctx.Err())
		return ErrDispatcherContextDone
	case d.unsubscribeChan <- c:
	}

	return nil
}

func (d *dispatcher[T]) Dispatch(ctx context.Context, t T) error {
	select {
	case <-d.ctx.Done():
		log.Errorf("Dispatcher context is done: %s", d.ctx.Err())
		return ErrDispatcherContextDone
	case <-ctx.Done():
		log.Errorf("Dispatch context is done: %s", ctx.Err())
		return ErrDispatchContextDone
	case d.dispatch <- t:
		return nil
	}
}

func (d *dispatcher[T]) Stop() {
	if d == nil {
		return
	}

	d.cancelFn()
}

func (d *dispatcher[T]) run() {
	if d == nil {
		return
	}

	subscribers := make(map[chan T]struct{})

	for {
		select {
		case <-d.ctx.Done():
			log.Errorf("Dispatcher context is done: %s", d.ctx.Err())
			return
		case c := <-d.subscribeChan:
			subscribers[c] = struct{}{}
		case c := <-d.unsubscribeChan:
			delete(subscribers, c)
			close(c)
		case p := <-d.dispatch:
			for subscriber := range subscribers {
				select {
				case <-d.ctx.Done():
					log.Errorf("Dispatcher context is done: %s", d.ctx.Err())
					return
				case <-time.After(d.opts.sendTimeout):
					log.Warn("Couldn't send message to subscriber")
					continue
				case subscriber <- p:
				}
			}
		}
	}
}
