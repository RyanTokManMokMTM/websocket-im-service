package demo

import (
	"github.com/ryantokmanmokmtm/websocket-im-service/server"
	"github.com/ryantokmanmokmtm/websocket-im-service/server/websocket"
	"log"
	"time"
)

type Server struct {
}

func (s *Server) StartServer(protocol, addr string, port uint) {
	//Server is depends on Protocol
	var ser server.Server
	if protocol == "TCP" {
		//USING TCP Server
		return
	} else if protocol == "WS" {
		ser = websocket.NewServer(addr, port)
	}

	handler := &ServerHandler{}
	//Server Setting
	ser.SetAcceptor(handler)
	ser.SetMessageListener(handler)
	ser.SetStateListener(handler)
	ser.SetReadWaitTime(time.Minute)

	err := ser.Start()
	if err != nil {
		panic(err)
	}
}

//Listener

type ServerHandler struct {
}

func (ser *ServerHandler) Accept(server.Conn, time.Duration) (string, error) {
	//FAKE ACCEPT
	//get user token?
	return "", nil
}
func (ser *ServerHandler) Receive(agent server.Agent, message []byte) {
	//msg := string(message) + ""
	_ = agent.Push(message)
}

func (ser *ServerHandler) Disconnect(id string) error {
	log.Printf("Client %s disconnet", id)
	return nil
}
