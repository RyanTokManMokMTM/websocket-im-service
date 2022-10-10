package tcp

import (
	"github.com/ryantokmanmokmtm/websocket-im-service/server"
	"github.com/ryantokmanmokmtm/websocket-im-service/util"
	"net"
)

type Frame struct {
	OpCode  server.OpCode
	Payload []byte
}

/*
  SetOpCode(code OpCode)
    GetOpCode() OpCode
    SetPayload([]byte)
    GetPayload() []byte
*/

func (f *Frame) SetOpCode(code server.OpCode) {
	f.OpCode = code
}
func (f *Frame) GetOpCode() server.OpCode {
	return f.OpCode
}
func (f *Frame) SetPayload(payload []byte) {
	f.Payload = payload
}
func (f *Frame) GetPayload() []byte {
	return f.Payload
}

type TcpConn struct {
	net.Conn
}

func NewTcpConn(conn net.Conn) *TcpConn {
	return &TcpConn{
		Conn: conn,
	}
}

//Following TCP Package Serialization
/*
Following this structure
Opcode which is a byte -> 8bit
Payload which is len + data
*/

func (t *TcpConn) ReadFrame() (server.Frame, error) {
	//Read OpCode
	opcode, err := util.ReadUint8(t.Conn)
	if err != nil {
		return nil, err
	}
	//Read data
	payload, err := util.ReadData(t.Conn)
	if err != nil {
		return nil, err
	}
	return &Frame{
		OpCode:  server.OpCode(opcode),
		Payload: payload,
	}, nil
}

func (t *TcpConn) WriteFrame(code server.OpCode, data []byte) error {
	//write opcode
	err := util.WriteUint8(t.Conn, uint8(code))
	if err != nil {
		return err
	}
	//write data
	err = util.WriteData(t.Conn, data)
	if err != nil {
		return err
	}

	return nil
}

func (t *TcpConn) Flush() error {
	return nil
}
