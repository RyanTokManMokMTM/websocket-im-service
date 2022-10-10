package tcp

import (
	"fmt"
	"github.com/ryantokmanmokmtm/websocket-im-service/server"
	"log"
	"net/url"
	"sync"
	"sync/atomic"
	"time"
)

type TcpClient struct {
	sync.Mutex
	server.Dialer
	once        sync.Once
	id          string
	name        string
	conn        server.Conn
	readWait    time.Duration
	writeWait   time.Duration
	heartBeat   time.Duration
	isConnected int32
}

func NewTcpClient(id, name string) server.Client {
	return &TcpClient{
		id:        id,
		name:      name,
		readWait:  server.DefaultReadWait,
		writeWait: server.DefaultWriteWait,
	}
}

func (c *TcpClient) ID() string {
	return c.id
}
func (c *TcpClient) Name() string {
	return c.name
}
func (c *TcpClient) Connect(addr string) error {
	_, err := url.Parse(addr)
	if err != nil {
		return err
	}

	if !atomic.CompareAndSwapInt32(&c.isConnected, 0, 1) {
		return fmt.Errorf("client is connected")
	}

	conn, err := c.DialAndHandShake(server.DialerContext{
		Id:      c.id,
		Name:    c.name,
		Addr:    addr,
		TimeOut: server.DefaultAcceptWait,
	})

	if err != nil {
		atomic.CompareAndSwapInt32(&c.isConnected, 1, 0)
		return err
	}

	//encapsulate conn as server.con(TCP Connection)
	//which conn implemented ReadFrame WriteFrame for TCP

	c.conn = NewTcpConn(conn)

	if c.heartBeat > 0 {
		go func() {
			err := c.heartBeatLoop()
			if err != nil {
				log.Printf("TCP Client heartBeat loop err: %v", err)
			}
		}()
	}

	return nil
}
func (c *TcpClient) SetDialer(dialer server.Dialer) {
	c.Dialer = dialer
}

func (c *TcpClient) Send(message []byte) error {
	//client still connected?
	if atomic.LoadInt32(&c.isConnected) == 0 {
		return fmt.Errorf("client connection is nil")
	}

	c.Lock()
	defer c.Unlock()

	if err := c.conn.SetWriteDeadline(time.Now().Add(c.writeWait)); err != nil {
		return err
	}

	return c.conn.WriteFrame(server.OpBinary, message)
}

func (c *TcpClient) Read() (server.Frame, error) {
	//check client is still connected
	if c.conn == nil {
		return nil, fmt.Errorf("client connection is nil")
	}

	err := c.conn.SetReadDeadline(time.Now().Add(c.readWait))
	if err != nil {
		return nil, err
	}

	f, err := c.conn.ReadFrame()
	if err != nil {
		return nil, err
	}

	if f.GetOpCode() == server.OpClose {
		return nil, fmt.Errorf("server closed the channel")
	}
	return f, nil
}

func (c *TcpClient) Close() {
	c.once.Do(func() {
		if c.conn == nil {
			return
		}

		//send message to server
		_ = c.conn.WriteFrame(server.OpClose, nil)
		c.conn.Close()
		atomic.CompareAndSwapInt32(&c.isConnected, 1, 0)
	})
}

func (c *TcpClient) heartBeatLoop() error {
	t := time.NewTicker(c.heartBeat)
	for {
		select {
		case <-t.C:
			log.Println("Client is sending a ping to server...")
			if err := c.Ping(); err != nil {
				return err
			}
		}
	}
}

func (c *TcpClient) Ping() error {
	//reset write deadline
	err := c.conn.SetWriteDeadline(time.Now().Add(c.writeWait))
	if err != nil {
		return err
	}

	return c.conn.WriteFrame(server.OpPing, nil)
}
