package channel_event

import (
	"sync"
	"sync/atomic"
)

type ChannelEvent struct {
	closed int32 //that represent channel is closed or not
	ch     chan struct{}
	once   sync.Once
}

func NewChannelEvent() *ChannelEvent {
	return &ChannelEvent{
		ch: make(chan struct{}),
	}
}

func (ce *ChannelEvent) Close() bool {
	//close the channel
	isClosed := false
	ce.once.Do(func() {
		atomic.StoreInt32(&ce.closed, 1) // Thread safety
		close(ce.ch)
		isClosed = true
	})
	return isClosed //yes closed
}

func (ce *ChannelEvent) Done() <-chan struct{} {
	return ce.ch // if channel is closed,it will send a sign.
}

func (ce *ChannelEvent) HasClosed() bool {
	return atomic.LoadInt32(&ce.closed) == 1
}
