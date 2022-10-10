package server

import (
	"context"
	"net"
	"time"
)

//Frame OpCode
type OpCode byte

const (
	OpContinuation OpCode = 0x0
	OpText         OpCode = 0x1
	OpBinary       OpCode = 0x2
	OpClose        OpCode = 0x8
	OpPing         OpCode = 0x9
	OpPong         OpCode = 0xa
)

const (
	DefaultReadWait   = time.Minute * 3
	DefaultWriteWait  = time.Second * 10
	DefaultAcceptWait = time.Second * 10
)

//Define Server Layer
type Server interface {
	SetAcceptor(acceptor Acceptor)               //set an acceptor to handle TPC/Websocket handshaking
	SetMessageListener(listener MessageListener) //set a message listener for read/write message
	SetStateListener(listener StateListener)     //set a state listener for handle connection state
	SetChannelMap(channelMap ChannelMap)         //set a channel map for manage user/agent
	SetReadWaitTime(duration time.Duration)      //set read message time for health check
	//Server Function
	Start() error                       //start the server
	Push(string, []byte) error          //push message to client/agent @params string - user_id, @params []byte - message
	Shutdown(ctx context.Context) error //close the server
}

type Acceptor interface {
	Accept(Conn, time.Duration) (string, error) //accept connection logic
}

type MessageListener interface {
	Receive(Agent, []byte)
}

type Agent interface {
	ID() string        //channelID/UserId?
	Push([]byte) error //push message to upper layer for handling the message
}

type StateListener interface {
	Disconnect(string) error //disconnect the connection with channelID/UserID
}

type ClientChannel interface {
	Conn
	Agent
	Close() error
	ReadLoop(listener MessageListener) error //this MessageListener.Receive is clientChannel(Agent) itself
	SetReadWait(duration time.Duration)
	SetWriteWait(duration time.Duration)
}

//Conn - TCP/Websocket connection
type Conn interface {
	net.Conn
	ReadFrame() (Frame, error)
	WriteFrame(OpCode, []byte) error
	Flush() error //What to do??
}

//TCP/Websocket Frame
type Frame interface {
	SetOpCode(code OpCode)
	GetOpCode() OpCode
	SetPayload([]byte) //server side need to decode payload before return
	GetPayload() []byte
}
