package websocket

import (
	"context"
	"errors"
	"fmt"
	"github.com/gobwas/ws"
	"github.com/ryantokmanmokmtm/websocket-im-service/server"
	"github.com/segmentio/ksuid"
	"log"
	"net/http"
	"sync"
	"time"
)

type WsServer struct {
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
}

func NewServer(addr string, port uint) server.Server {
	return &WsServer{
		addr:       addr,
		port:       port,
		readWait:   server.DefaultReadWait,
		writeWait:  server.DefaultReadWait,
		acceptWait: server.DefaultAcceptWait,
	}
}

/*
	SetAcceptor(acceptor Acceptor)               //set an acceptor to handle TPC/Websocket handshaking
	SetMessageListener(listener MessageListener) //set a message listener for read/write message
	SetStateListener(listener StateListener)     //set a state listener for handle connection state
	SetChannelMap(channelMap ChannelMap)         //set a channel map for manage user/agent
	SetReadWaitTime(duration time.Duration)      //set read message time for health check
	//Server Function
	Start() error                       //start the server
	Push(string, []byte) error               //push message to client/agent @params string - user_id, @params []byte - message
	Shutdown(ctx context.Context) error //close the server
*/

func (s *WsServer) Start() error {
	log.Printf("Initing Websocket server at %v:%v", s.addr, s.port)
	if s.Acceptor == nil {
		s.Acceptor = new(DefaultAcceptor) //if not set
	}

	if s.StateListener == nil {
		return fmt.Errorf("state listener is nil")
	}

	if s.MessageListener == nil {
		return fmt.Errorf("message listener is nil")
	}

	if s.ChannelMap == nil {
		s.ChannelMap = server.NewChannelsMap()
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		//Upgrade http to websocket
		log.Printf("Upgrad HTTP To Websocket")
		conn, _, _, err := ws.UpgradeHTTP(r, w)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(err.Error()))
		}

		//Make own connection object
		wsConn := NewWsConn(conn)

		//handshaking - user connected
		id, err := s.Accept(wsConn, s.acceptWait) //Generate Global Unique ID
		if err != nil {
			wsConn.WriteFrame(server.OpClose, []byte(err.Error())) //Write a close frame to client
			wsConn.Close()                                         //close connection
			return
		}

		//manager id from channelMap
		if _, ok := s.Get(id); ok {
			log.Printf("channel id %s existed", id)
			wsConn.WriteFrame(server.OpClose, []byte("channel id existed!"))
			wsConn.Close() //close connection
			return
		}

		//create clientChannel
		client := server.NewClientChannelImpl(id, wsConn)
		//client setting
		client.SetReadWait(s.readWait)
		client.SetWriteWait(s.writeWait)
		//insert into channel map
		s.Add(client)

		//create a goroutine from client to read data
		go func(client server.ClientChannel) {
			err := client.ReadLoop(s.MessageListener)
			if err != nil {
				log.Printf("read loop error(id - %s) : %s", client.ID(), err.Error())
			}

			//disconnect the connection of the client
			//remove client from the channel mpa
			s.Remove(client.ID())
			err = s.Disconnect(client.ID()) //used to announce upper layer to handle
			if err != nil {
				log.Printf("disconnectd state error %s", err.Error())
			}

			client.Close()
		}(client)
	})
	log.Println("Websocket Server Started!")
	return http.ListenAndServe(fmt.Sprintf("%s:%d", s.addr, s.port), mux)
}

func (s *WsServer) Push(id string, message []byte) error {
	//push message to client
	ch, ok := s.ChannelMap.Get(id)
	if !ok {
		return errors.New("client not found")
	}
	return ch.Push(message)
}

func (s *WsServer) Shutdown(ctx context.Context) error {
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

func (s *WsServer) SetAcceptor(acceptor server.Acceptor) {
	s.Acceptor = acceptor
}

func (s *WsServer) SetMessageListener(messageListener server.MessageListener) {
	s.MessageListener = messageListener
}

func (s *WsServer) SetStateListener(stateListener server.StateListener) {
	s.StateListener = stateListener
}

func (s *WsServer) SetChannelMap(channelMap server.ChannelMap) {
	s.ChannelMap = channelMap
}

func (s *WsServer) SetReadWaitTime(t time.Duration) {
	s.readWait = t
}

//DefaultAcceptor - Default Acceptor
type DefaultAcceptor struct {
}

func (a *DefaultAcceptor) Accept(conn server.Conn, t time.Duration) (string, error) {
	//SIMPLY RETURN A UNIQUE ID AS CHANNEL ID - USING `KSUID`
	return ksuid.New().String(), nil
}
