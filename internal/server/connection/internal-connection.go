package connection

import (
	"fmt"
	"io"
	"log"
	"net"
	"proxy/internal/server/messagehandler"
	"proxy/internal/server/pack"
	"proxy/pkg/message"

	"github.com/hashicorp/go-uuid"
	goeh "github.com/hetacode/go-eh"
)

type InternalConnection struct {
	ID                   string
	Host                 string
	connection           net.Conn
	chanRemoveConnection chan<- string
	chanAddConnection    chan<- pack.ChanInternalConnection
	eventsMapper         *goeh.EventsMapper
	eventsHandlerManager *goeh.EventsHandlerManager
}

func NewInternalConnection(con net.Conn, chanRemoveConnection chan<- string, chanAddConnection chan<- pack.ChanInternalConnection) *InternalConnection {
	id, _ := uuid.GenerateUUID()
	c := &InternalConnection{
		ID:                   id,
		connection:           con,
		chanRemoveConnection: chanRemoveConnection,
		chanAddConnection:    chanAddConnection,
		eventsMapper:         message.NewEventsMapper(),
	}
	c.eventsHandlerManager = c.registerMessageHandlers()
	return c
}

func (c *InternalConnection) Send(msgBytes []byte) error {
	if c.connection == nil {
		return fmt.Errorf("connection is not initialized")
	}

	_, err := c.connection.Write(msgBytes)
	if err != nil {
		return err
	}

	return nil
}

func (c *InternalConnection) Listen() {
	log.Printf("New connection: %s", c.connection.RemoteAddr().String())
	msgBytes := make([]byte, 0)
	for {
		b := make([]byte, 1024)
		bl, err := c.connection.Read(b)
		if err != nil {
			if err == io.EOF {
				c.parseBytesMessage(msgBytes)
				msgBytes = make([]byte, 0)
				continue
			}
			c.chanRemoveConnection <- c.Host
			log.Print(err)
			break
		}
		log.Printf("Recv bytes: %d", bl)
		msgBytes = append(msgBytes, b[:bl]...)
		if bl < len(b) {
			c.parseBytesMessage(msgBytes)
			msgBytes = make([]byte, 0)
			continue
		}
	}
	log.Printf("End receiving")
	c.chanRemoveConnection <- c.Host
}

func (c *InternalConnection) SetHost(host string) {
	c.Host = host
	c.chanAddConnection <- pack.ChanInternalConnection{
		Host:       host,
		Connection: c,
	}
}

func (c *InternalConnection) parseBytesMessage(msgBytes []byte) {
	event, err := c.eventsMapper.Resolve(string(msgBytes))
	if err != nil {
		log.Print(err)
		return
	}
	c.eventsHandlerManager.Execute(event)
}

func (c *InternalConnection) registerMessageHandlers() *goeh.EventsHandlerManager {
	ehm := goeh.NewEventsHandlerManager()
	ehm.Register(new(message.AgentRegistrationMessage), &messagehandler.AgentRegistrationMessageHandler{Connection: c})
	return ehm

}
