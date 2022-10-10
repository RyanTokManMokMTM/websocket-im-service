package websocket

import (
	"fmt"
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"github.com/ryantokmanmokmtm/websocket-im-service/server"
	"log"
	"net"
	"net/url"
	"sync"
	"sync/atomic"
	"time"
)

type WsClient struct {
	sync.Mutex
	server.Dialer
	once        sync.Once
	id          string
	name        string
	conn        net.Conn
	readWait    time.Duration
	writeWait   time.Duration
	heartBeat   time.Duration
	IsConnected int32 //represent user is connected or not
	ctx         *server.DialerContext
}

/*
	ID() string           //Client ID
	Name() string         //Client Name
	Connect(string) error //Connect to server

	SetDialer(Dialer)
	Send([]byte) error
	Read() (server.Frame, error)
	Close()

*/
func NewClient(id, name string) server.Client {
	return &WsClient{
		id:        id,
		name:      name,
		readWait:  server.DefaultReadWait,
		writeWait: server.DefaultWriteWait,
	}
}

func (c *WsClient) ID() string {
	return c.id
}
func (c *WsClient) Name() string {
	return c.name
}
func (c *WsClient) Connect(addr string) error {
	_, err := url.Parse(addr)
	if err != nil {
		return err
	}

	//swap 1 and 1 -> false
	//swap 0 and 1 -> true
	if !atomic.CompareAndSwapInt32(&c.IsConnected, 0, 1) {
		return fmt.Errorf("client has connected")
	}

	//Handshaking - handle in upper layer
	conn, err := c.Dialer.DialAndHandShake(server.DialerContext{
		Id:      c.id,
		Name:    c.name,
		Addr:    addr,
		TimeOut: server.DefaultAcceptWait,
	})

	if err != nil {
		//change connected state back to 0
		atomic.CompareAndSwapInt32(&c.IsConnected, 1, 0)
		return err
	}

	//check connection
	if conn == nil {
		return fmt.Errorf("connection is nil")
	}

	//need health check?
	if c.heartBeat > 0 {
		//create a new goroutine for sending ping/heartBeart
		go func() {
			err := c.heartBeatLoop()
			if err != nil {
				log.Println("heartbeat loop error")
			}
		}()
	}

	return nil
}
func (c *WsClient) SetDialer(dia server.Dialer) {
	c.Dialer = dia
}

func (c *WsClient) Send(message []byte) error {
	//client still connected?
	if atomic.LoadInt32(&c.IsConnected) == 0 {
		return fmt.Errorf("client connect is nil")
	}

	c.Mutex.Lock()
	defer c.Mutex.Unlock()
	//reset the deadline time
	err := c.conn.SetWriteDeadline(time.Now().Add(c.writeWait))
	if err != nil {
		return err
	}

	//send message with mask
	return wsutil.WriteClientMessage(c.conn, ws.OpBinary, message)
}
func (c *WsClient) Read() (server.Frame, error) {
	//check client is still connected
	if c.conn == nil {
		return nil, fmt.Errorf("client connection is nil")
	}

	//reset read deadline
	if c.heartBeat > 0 {
		_ = c.conn.SetReadDeadline(time.Now().Add(c.readWait))
	}

	f, err := ws.ReadFrame(c.conn)
	if err != nil {
		return nil, err
	}

	if f.Header.OpCode == ws.OpClose {
		return nil, fmt.Errorf("server close the channel")
	}

	return &WsFrame{
		raw: f,
	}, nil
}
func (c *WsClient) Close() {
	c.once.Do(func() {
		if c.conn == nil {
			return
		}

		//send a message to server
		_ = wsutil.WriteClientMessage(c.conn, ws.OpClose, nil)
		//close the connection
		c.conn.Close()
		//switch client state
		atomic.CompareAndSwapInt32(&c.IsConnected, 1, 0)
	})
}

func (c *WsClient) heartBeatLoop() error {
	//set a timer to send a heart bear
	//ticker time > write deadline
	t := time.NewTicker(c.heartBeat)
	select {
	case <-t.C:
		//send a ping to server
		log.Println("Is time to send a ping to server...")
		if err := c.Ping(); err != nil {
			return err
		}
	}
	return nil
}

func (c *WsClient) Ping() error {
	c.Lock()
	defer c.Unlock()

	if err := c.conn.SetWriteDeadline(time.Now().Add(c.writeWait)); err != nil {
		return err
	}

	log.Printf("client %s send a ping to server\n", c.id)
	return wsutil.WriteClientMessage(c.conn, ws.OpPing, nil)
}
