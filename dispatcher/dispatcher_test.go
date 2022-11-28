//go:build unit

package dispatcher

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewDispatcher(t *testing.T) {
	ctx, canc := context.WithCancel(context.Background())
	defer canc()

	chanBuf := 11
	sendTimeout := 5 * time.Second

	disp := NewDispatcher[string](
		ctx,
		WithDispatchChanBuf(chanBuf),
		WithSubscribeChanBuf(chanBuf),
		WithUnsubscribeChanBuf(chanBuf),
		WithSubscriberChanBuf(chanBuf),
		WithSendTimeout(sendTimeout),
	)

	defer disp.Stop()

	testDisp := disp.(*dispatcher[string])
	assert.Equal(t, chanBuf, testDisp.opts.dispatchChanBuf)
	assert.Equal(t, chanBuf, testDisp.opts.subscribeChanBuf)
	assert.Equal(t, chanBuf, testDisp.opts.unsubscribeChanBuf)
	assert.Equal(t, chanBuf, testDisp.opts.subscriberChanBuf)
	assert.Equal(t, sendTimeout, testDisp.opts.sendTimeout)
}

func TestDispatcher_SubscribeAndDispatch(t *testing.T) {
	ctx, canc := context.WithCancel(context.Background())
	defer canc()

	dispatcher := NewDispatcher[string](ctx)

	c, err := dispatcher.Subscribe(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, c)

	testString := "test"

	time.Sleep(1 * time.Second)

	err = dispatcher.Dispatch(ctx, testString)
	assert.NoError(t, err)

	select {
	case res := <-c:
		assert.Equal(t, testString, res)
		return
	case <-time.After(3 * time.Second):
		assert.Fail(t, "dispatched message was not received")
	}
}

func TestDispatcher_Subscribe_ContextErrors(t *testing.T) {
	ctx, canc := context.WithCancel(context.Background())

	dispatcher := NewDispatcher[string](ctx, WithSubscribeChanBuf(0))

	canc()

	c, err := dispatcher.Subscribe(context.Background())
	assert.ErrorIs(t, err, ErrDispatcherContextDone)
	assert.Nil(t, c)

	ctx, canc = context.WithCancel(context.Background())
	defer canc()

	dispatcher = NewDispatcher[string](ctx, WithSubscribeChanBuf(0))

	subscribeCtx, subscribeCtxCanc := context.WithCancel(context.Background())

	subscribeCtxCanc()

	c, err = dispatcher.Subscribe(subscribeCtx)
	assert.ErrorIs(t, err, ErrSubscribeContextDone)
	assert.Nil(t, c)
}

func TestDispatcher_Unsubscribe(t *testing.T) {
	ctx, canc := context.WithCancel(context.Background())
	defer canc()

	dispatcher := NewDispatcher[string](
		ctx,
		WithSubscribeChanBuf(0),
		WithUnsubscribeChanBuf(0),
	)

	c, err := dispatcher.Subscribe(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, c)

	err = dispatcher.Unsubscribe(c)
	assert.NoError(t, err)

	select {
	case _, ok := <-c:
		assert.False(t, ok)
		return
	case <-time.After(3 * time.Second):
		assert.Fail(t, "subscriber channel not closed")
	}
}

func TestDispatcher_Unsubscribe_ContextErrors(t *testing.T) {
	ctx, canc := context.WithCancel(context.Background())

	dispatcher := NewDispatcher[string](
		ctx,
		WithSubscribeChanBuf(0),
		WithUnsubscribeChanBuf(0),
	)

	c, err := dispatcher.Subscribe(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, c)

	canc()

	err = dispatcher.Unsubscribe(c)
	assert.ErrorIs(t, err, ErrDispatcherContextDone)
}

func TestDispatcher_Dispatch_ContextErrors(t *testing.T) {
	ctx, canc := context.WithCancel(context.Background())

	dispatcher := NewDispatcher[string](
		ctx,
		WithDispatchChanBuf(0),
	)

	canc()

	err := dispatcher.Dispatch(context.Background(), "test")
	assert.ErrorIs(t, err, ErrDispatcherContextDone)

	dispatcher = NewDispatcher[string](
		context.Background(),
		WithDispatchChanBuf(0),
	)

	dispatchCtx, dispatchCtxCanc := context.WithCancel(context.Background())

	dispatchCtxCanc()

	err = dispatcher.Dispatch(dispatchCtx, "test")
	assert.ErrorIs(t, err, ErrDispatchContextDone)
}

func TestDispatcher_Stop(t *testing.T) {
	dispatcher := NewDispatcher[string](
		context.Background(),
		WithDispatchChanBuf(0),
		WithSubscribeChanBuf(0),
		WithUnsubscribeChanBuf(0),
	)

	dispatcher.Stop()

	c, err := dispatcher.Subscribe(context.Background())
	assert.ErrorIs(t, err, ErrDispatcherContextDone)
	assert.Nil(t, c)

	err = dispatcher.Unsubscribe(c)
	assert.ErrorIs(t, err, ErrDispatcherContextDone)

	err = dispatcher.Dispatch(context.Background(), "test")
	assert.ErrorIs(t, err, ErrDispatcherContextDone)
}
