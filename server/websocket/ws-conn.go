package websocket

import (
	"github.com/gobwas/ws"
	"github.com/ryantokmanmokmtm/websocket-im-service/server"
	"net"
)

type WsFrame struct {
	raw ws.Frame //Websocket frame
}

/*
	SetOpCode(code OpCode)
	GetOpCode() OpCode
	SetPayload([]byte) //server side need to decode payload before return
	GetPayload() []byte
*/

func (f *WsFrame) SetOpCode(code server.OpCode) {}

func (f *WsFrame) GetOpCode() server.OpCode {
	return server.OpCode(f.raw.Header.OpCode)
}

func (f *WsFrame) SetPayload(payload []byte) {}

func (f *WsFrame) GetPayload() []byte {
	return nil
}

/*
	net.Conn
	ReadFrame() (Frame, error)
	WriteFrame(OpCode, []byte) error
	Flush() error //What to do??
*/

type WsConn struct {
	net.Conn
}

func NewWsConn(conn net.Conn) *WsConn {
	return &WsConn{
		Conn: conn,
	}
}

func (c *WsConn) ReadFrame() (server.Frame, error) {
	f, err := ws.ReadFrame(c.Conn)
	if err != nil {
		return nil, err
	}

	return &WsFrame{raw: f}, nil
}

func (c *WsConn) WriteFrame(code server.OpCode, message []byte) error {
	//New a frame
	f := ws.NewFrame(ws.OpCode(code), true, message)
	return ws.WriteFrame(c.Conn, f)
}

func (c *WsConn) Flush() error {
	return nil
}
