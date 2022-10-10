package server

import (
	"net"
	"time"
)

type Client interface {
	ID() string           //Client ID
	Name() string         //Client Name
	Connect(string) error //Connect to server

	SetDialer(Dialer)
	Send([]byte) error
	Read() (Frame, error)
	Close()
}

type Dialer interface {
	DialAndHandShake(DialerContext) (net.Conn, error)
}

type DialerContext struct {
	Id      string
	Name    string
	Addr    string
	TimeOut time.Duration
}
