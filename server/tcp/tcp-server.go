package tcp

import (
	"context"
	"errors"
	"fmt"
	channel_event "github.com/ryantokmanmokmtm/websocket-im-service/channel-event"
	"github.com/ryantokmanmokmtm/websocket-im-service/server"
	"github.com/segmentio/ksuid"
	"log"
	"net"
	"sync"
	"time"
)

type TcpServer struct {
	addr string
	port uint
	server.ChannelMap
	server.Acceptor
	server.MessageListener
	server.StateListener
	once       sync.Once
	readWait   time.Duration
	writeWait  time.Duration
	acceptWait time.Duration
	close      *channel_event.ChannelEvent
}

func NewTcpServer(addr string, port uint) *TcpServer {
	return &TcpServer{
		addr:       addr,
		port:       port,
		ChannelMap: server.NewChannelsMap(),
		readWait:   server.DefaultReadWait,
		writeWait:  server.DefaultWriteWait,
		acceptWait: server.DefaultAcceptWait,
		close:      channel_event.NewChannelEvent(),
	}
}

//Implement Server Interface

func (s *TcpServer) Start() error {
	log.Printf("Initing TCP server at %v:%v", s.addr, s.port)

	if s.Acceptor == nil {
		s.Acceptor = new(DefaultAcceptor) //if not set
	}

	if s.StateListener == nil {
		return fmt.Errorf("state listener is nil")
	}

	if s.MessageListener == nil {
		return fmt.Errorf("message listener is nil")
	}
	//
	//if s.ChannelMap == nil {
	//	s.ChannelMap = server.NewChannelsMap()
	//}

	tcp, err := net.Listen("tcp", fmt.Sprintf("%s:%d", s.addr, s.port))
	if err != nil {
		return err
	}

	log.Println("server is up!")
	for {
		//accept connection
		conn, err := tcp.Accept() //For TCP Server
		if err != nil {
			//close the connection
			conn.Close()
			log.Println(err)
			continue
		}

		go func(conn net.Conn) {
			//create conn object
			newConn := NewTcpConn(conn)

			//accept
			id, err := s.Accept(newConn, s.acceptWait)
			if err != nil {
				log.Println("accept connection error")
				_ = newConn.WriteFrame(server.OpClose, nil)
				newConn.Close()
				return
			}

			//check id
			if _, ok := s.Get(id); ok {
				log.Println("Client Channel ID existed!")
				_ = newConn.WriteFrame(server.OpClose, []byte("Client Channel ID existed"))
				newConn.Close()
				return
			}

			//Setting Client
			client := server.NewClientChannelImpl(id, newConn)
			client.SetReadWait(s.readWait)
			client.SetWriteWait(s.writeWait)
			//add to map
			s.Add(client)

			//read loop
			err = client.ReadLoop(s.MessageListener)
			if err != nil {
				log.Println(err)
			}

			//remove client from map
			s.Remove(id)

			//disconnect connection
			_ = s.Disconnect(id)

			//close connection
			newConn.Close()
		}(conn)

		select {
		case <-s.close.Done(): //Where is this signal come from shutdown?
			return fmt.Errorf("tcp server existed")

		default:
		}
	}
}

func (s *TcpServer) Push(id string, message []byte) error {
	//push message to client
	ch, ok := s.ChannelMap.Get(id)
	if !ok {
		return errors.New("client not found")
	}
	return ch.Push(message)
}

func (s *TcpServer) Shutdown(ctx context.Context) error {
	log.Println("Server is shutting down...")
	s.once.Do(func() {
		defer func() {
			log.Printf("shut down!")
		}()
		for _, client := range s.ChannelMap.All() {
			client.Close()
			select {
			case <-ctx.Done(): //stop immediately
				return
			default:
				continue
			}
		}
	})
	return nil
}

func (s *TcpServer) SetAcceptor(acceptor server.Acceptor) {
	s.Acceptor = acceptor
}

func (s *TcpServer) SetMessageListener(messageListener server.MessageListener) {
	s.MessageListener = messageListener
}

func (s *TcpServer) SetStateListener(stateListener server.StateListener) {
	s.StateListener = stateListener
}

func (s *TcpServer) SetChannelMap(channelMap server.ChannelMap) {
	s.ChannelMap = channelMap
}

func (s *TcpServer) SetReadWaitTime(t time.Duration) {
	s.readWait = t
}

//DefaultAcceptor - Default Acceptor
type DefaultAcceptor struct {
}

func (a *DefaultAcceptor) Accept(conn server.Conn, t time.Duration) (string, error) {
	//SIMPLY RETURN A UNIQUE ID AS CHANNEL ID - USING `KSUID`
	return ksuid.New().String(), nil
}
