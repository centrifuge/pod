package dispatcher

import "time"

type opts struct {
	dispatchChanBuf    int
	subscribeChanBuf   int
	unsubscribeChanBuf int
	subscriberChanBuf  int
	sendTimeout        time.Duration
}

const (
	defaultSendTimeout = 1 * time.Second
	defaultChanBuffer  = 10
)

func getDefaultOpts() *opts {
	return &opts{
		dispatchChanBuf:    defaultChanBuffer,
		subscribeChanBuf:   defaultChanBuffer,
		unsubscribeChanBuf: defaultChanBuffer,
		subscriberChanBuf:  defaultChanBuffer,
		sendTimeout:        defaultSendTimeout,
	}
}

type Opt func(o *opts)

func WithDispatchChanBuf(buf int) Opt {
	return func(o *opts) {
		o.dispatchChanBuf = buf
	}
}

func WithSubscribeChanBuf(buf int) Opt {
	return func(o *opts) {
		o.subscribeChanBuf = buf
	}
}

func WithUnsubscribeChanBuf(buf int) Opt {
	return func(o *opts) {
		o.unsubscribeChanBuf = buf
	}
}

func WithSubscriberChanBuf(buf int) Opt {
	return func(o *opts) {
		o.subscriberChanBuf = buf
	}
}

func WithSendTimeout(timeout time.Duration) Opt {
	return func(o *opts) {
		o.sendTimeout = timeout
	}
}
