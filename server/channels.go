package server

import (
	"errors"
	"fmt"
	channel_event "github.com/ryantokmanmokmtm/websocket-im-service/channel-event"
	"log"
	"sync"
	"time"
)

// ClientChannelImpl - A ClientChannel Implementation of Websocket
type ClientChannelImpl struct {
	sync.Mutex
	id string
	Conn
	writeChan chan []byte
	once      sync.Once
	writeWait time.Duration
	readWait  time.Duration
	close     *channel_event.ChannelEvent
}

/*
type ClientChannel interface {
	Conn
	Agent
	Close() error
	ReadLoop(listener MessageListener) error //this MessageListener.Receive is clientChannel(Agent) itself
	SetReadWait(duration time.Duration)
	SetWriteWait(duration time.Duration)
}
*/

func NewClientChannelImpl(id string, conn Conn) *ClientChannelImpl {
	log.Printf("Client Channel with id %v", id)
	ch := &ClientChannelImpl{
		id:        id,
		Conn:      conn,
		writeChan: make(chan []byte, 5), //??
		writeWait: time.Second * 10,
		readWait:  time.Second * 10,
		close:     channel_event.NewChannelEvent(),
	}

	//new a goroutine for write
	go func() {
		err := ch.writeLoop() //write
		if err != nil {
			log.Println(err)
		}
	}()

	return ch
}

func (chImp *ClientChannelImpl) writeLoop() error {
	for {
		select {
		case data := <-chImp.writeChan:
			//any payload from under layer ? Agent push to here
			//1. check opcode
			//2. send
			err := chImp.WriteFrame(OpBinary, data)
			if err != nil {
				log.Printf("write frame error : %v", err)
				return err
			}

			len := len(chImp.writeChan)
			log.Println("Channel Write Len", len)
			log.Println("Send reminding data")
			for i := 0; i < len; i++ {
				payLoad := <-chImp.writeChan
				err := chImp.WriteFrame(OpBinary, payLoad)
				if err != nil {
					log.Printf("write frame error : %v", err)
					return err
				}
				err = chImp.Conn.Flush() //What to Do??? TODO: Clear all data??
				if err != nil {
					return err
				}
			}
		case <-chImp.close.Done():
			log.Println("Write Channel has been closed!")
			return nil
		}
	}
}

func (chImp *ClientChannelImpl) ID() string {
	//implement Agent interface
	return chImp.id
}

func (chImp *ClientChannelImpl) Push(payload []byte) error {
	//channel has been closed?
	if chImp.close.HasClosed() {
		return fmt.Errorf("client (%s)'s Write Channel has been closed", chImp.id)
	}
	chImp.writeChan <- payload //push message to client channel to write message to client
	return nil
}

func (chImp *ClientChannelImpl) WriteFrame(code OpCode, bytes []byte) error {
	//resize write deadline
	_ = chImp.Conn.SetWriteDeadline(time.Now().Add(chImp.writeWait))
	return chImp.Conn.WriteFrame(code, bytes)
}

//Implment client channel interface
func (chImp *ClientChannelImpl) Close() error {
	//close the channel
	chImp.once.Do(func() {
		close(chImp.writeChan)
		//TODO: set some event here and let upper layer know the channel is closed!
	})
	return nil
}

func (chImp *ClientChannelImpl) ReadLoop(listener MessageListener) error {
	//Avoid to run read loop at different routine
	chImp.Lock()
	defer chImp.Unlock()
	for {
		//resize the read deadline
		_ = chImp.SetReadDeadline(time.Now().Add(chImp.readWait))
		f, err := chImp.ReadFrame()
		if err != nil {
			return err
		}

		if f.GetOpCode() == OpClose {
			log.Println("Client Closed the connection")
			return errors.New("client side closed the connection")
		}

		if f.GetOpCode() == OpPing {
			log.Println("received an ping request from client.")
			_ = chImp.WriteFrame(OpPong, nil)
			continue
		}

		//get Payload
		payload := f.GetPayload()
		if len(payload) == 0 {
			log.Println("payload is empty.")
			continue
		}

		//send the payload to message listener(upper layer) for handing the message
		go listener.Receive(chImp, payload) //clientChannel is an agent which is implemented Agent Interface -- TODO
	}
	return nil
}

func (chImp *ClientChannelImpl) SetReadWait(d time.Duration) {
	if d == 0 {
		return
	}
	chImp.readWait = d
}

func (chImp *ClientChannelImpl) SetWriteWait(d time.Duration) {
	if d == 0 {
		return
	}
	chImp.writeWait = d
}
