package connection

import (
	"fmt"
	"io"
	"log"
	"net"
	"proxy/internal/server/messagehandler"
	"proxy/internal/server/pack"
	"proxy/pkg/communication"
	"proxy/pkg/key"
	"proxy/pkg/message"
	"sync"

	"github.com/hashicorp/go-uuid"
	goeh "github.com/hetacode/go-eh"
)

type InternalConnection struct {
	ID                   string
	Host                 string
	connection           net.Conn
	chanRemoveConnection chan<- string
	chanAddConnection    chan<- pack.ChanInternalConnection
	chanMsgToExternal    chan<- pack.ChanProxyMessageToExternal
	eventsMapper         *goeh.EventsMapper
	eventsHandlerManager *goeh.EventsHandlerManager
	mutexSendMessage     sync.Mutex
}

func NewInternalConnection(con net.Conn, chanRemoveConnection chan<- string, chanAddConnection chan<- pack.ChanInternalConnection, chanMsgToExternal chan<- pack.ChanProxyMessageToExternal) *InternalConnection {
	id, _ := uuid.GenerateUUID()
	c := &InternalConnection{
		ID:                   id,
		connection:           con,
		chanRemoveConnection: chanRemoveConnection,
		chanAddConnection:    chanAddConnection,
		chanMsgToExternal:    chanMsgToExternal,
		eventsMapper:         message.NewEventsMapper(),
		mutexSendMessage:     sync.Mutex{},
	}
	c.eventsHandlerManager = c.registerMessageHandlers()
	return c
}

func (c *InternalConnection) Send(externalConnectionID string, msgBytes []byte) error {
	if c.connection == nil {
		return fmt.Errorf("internal connection is not initialized")
	}
	c.mutexSendMessage.Lock()
	defer c.mutexSendMessage.Unlock()

	headers := communication.BytesHeader{
		// This header should back from agent
		// Purpose of it is to return response to the correct external connection
		key.ExternalConnectionIDKey: externalConnectionID,
	}
	msgBytes = communication.SerializeBytesMessage(headers, msgBytes)
	_, err := c.connection.Write(msgBytes)
	if err != nil {
		return err
	}

	return nil
}

func (c *InternalConnection) Listen() {
	log.Printf("New internal connection: %s", c.connection.RemoteAddr().String())
	msgBytes := make([]byte, 0)
	for {
		b := make([]byte, 1024)
		bl, err := c.connection.Read(b)
		if err != nil {
			if err == io.EOF {
				msgBytes = make([]byte, 0)
				break
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
	c.connection.Close()
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
	headers, msgBytes := communication.DeserializeBytesMessage(msgBytes)

	// Case when the message should be forward to the external connection
	if externalConnectionID, ok := headers[key.ExternalConnectionIDKey]; ok {
		log.Printf("response for externalnConnectionId: %s", externalConnectionID)
		msg := pack.ChanProxyMessageToExternal{
			ExternalConnectionID: externalConnectionID,
			Content:              msgBytes,
		}
		c.chanMsgToExternal <- msg
	} else {
		// TODO: other implementation of messaging - base on BytesMessage and headers
		event, err := c.eventsMapper.Resolve(string(msgBytes))
		if err != nil {
			log.Print(err)
			return
		}
		c.eventsHandlerManager.Execute(event)
	}
}

func (c *InternalConnection) registerMessageHandlers() *goeh.EventsHandlerManager {
	ehm := goeh.NewEventsHandlerManager()
	ehm.Register(new(message.AgentRegistrationMessage), &messagehandler.AgentRegistrationMessageHandler{Connection: c})
	return ehm

}
