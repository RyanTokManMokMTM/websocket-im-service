package server

import (
	"errors"
	"sync"
)

type ChannelMap interface {
	Add(channel ClientChannel) error
	Remove(string)
	Get(string) (ClientChannel, bool)
	All() []ClientChannel
}

type ChannelMapImpl struct {
	channelsMap *sync.Map //a sync map - thread safety
}

func NewChannelsMap() *ChannelMapImpl {
	return &ChannelMapImpl{
		channelsMap: new(sync.Map),
	}
}

func (ch *ChannelMapImpl) Add(channel ClientChannel) error {
	//Client ID : ClientChannel
	if channel.ID() == "" {
		return errors.New("client Channel ID is empty")
	}
	ch.channelsMap.Store(channel.ID(), channel)
	return nil
}

func (ch *ChannelMapImpl) Remove(clientID string) {
	if clientID == "" {
		return
	}
	ch.channelsMap.Delete(clientID)
}

func (ch *ChannelMapImpl) Get(clientID string) (ClientChannel, bool) {
	channel, ok := ch.channelsMap.Load(clientID)
	if !ok {
		return nil, false
	}
	return channel.(ClientChannel), true
}

func (ch *ChannelMapImpl) All() []ClientChannel {
	channels := make([]ClientChannel, 0)
	ch.channelsMap.Range(func(key, value any) bool {
		channels = append(channels, value.(ClientChannel))
		return true
	})
	return channels
}
