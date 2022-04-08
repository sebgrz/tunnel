package connection

import (
	"fmt"
	"io"
	"log"
	"net"
	"proxy/internal/server/messagehandler"
	"proxy/internal/server/pack"
	"proxy/pkg/communication"
	"proxy/pkg/enum"
	"proxy/pkg/key"
	"proxy/pkg/message"
	"sync"

	"github.com/hashicorp/go-uuid"
	goeh "github.com/hetacode/go-eh"
)

type InternalConnection struct {
	ID                          string
	Host                        string
	Type                        enum.AgentConnectionType
	connection                  net.Conn
	chanRemoveConnection        chan<- pack.ChanInternalConnectionSpec
	chanCloseExternalConnection chan<- string
	chanAddConnection           chan<- pack.ChanInternalConnection
	chanMsgToExternal           chan<- pack.ChanProxyMessageToExternal
	eventsMapper                *goeh.EventsMapper
	eventsHandlerManager        *goeh.EventsHandlerManager
	mutexSendMessage            sync.Mutex
}

func NewInternalConnection(con net.Conn, chanRemoveConnection chan<- pack.ChanInternalConnectionSpec, chanCloseExternalConnection chan<- string, chanAddConnection chan<- pack.ChanInternalConnection, chanMsgToExternal chan<- pack.ChanProxyMessageToExternal) *InternalConnection {
	id, _ := uuid.GenerateUUID()
	c := &InternalConnection{
		ID:                          id,
		connection:                  con,
		chanRemoveConnection:        chanRemoveConnection,
		chanCloseExternalConnection: chanCloseExternalConnection,
		chanAddConnection:           chanAddConnection,
		chanMsgToExternal:           chanMsgToExternal,
		eventsMapper:                message.NewEventsMapper(),
		mutexSendMessage:            sync.Mutex{},
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
			c.chanRemoveConnection <- pack.ChanInternalConnectionSpec{
				Host:           c.Host,
				ConnectionType: c.Type,
			}
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
	c.chanRemoveConnection <- pack.ChanInternalConnectionSpec{
		Host:           c.Host,
		ConnectionType: c.Type,
	}
}

func (c *InternalConnection) SendWithHeaders(externalConnectionID string, headers communication.BytesHeader, msgBytes []byte) error {
	if c.connection == nil {
		return fmt.Errorf("internal connection is not initialized")
	}
	if headers == nil {
		return fmt.Errorf("headers parameter is nil")
	}

	c.mutexSendMessage.Lock()
	defer c.mutexSendMessage.Unlock()

	headers[key.ExternalConnectionIDKey] = externalConnectionID

	msgBytes = communication.SerializeBytesMessage(headers, msgBytes)
	_, err := c.connection.Write(msgBytes)
	if err != nil {
		return err
	}

	return nil
}

func (c *InternalConnection) Configure(hostname string, connectionType enum.AgentConnectionType) {
	c.Host = hostname
	c.Type = connectionType
	c.chanAddConnection <- pack.ChanInternalConnection{
		Host:           hostname,
		ConnectionType: connectionType,
		Connection:     c,
	}
}

func (c *InternalConnection) parseBytesMessage(msgBytes []byte) {
	headers, msgBytes := communication.DeserializeBytesMessage(msgBytes)

	if messageType, ok := headers[key.MessageTypeBytesHeader]; ok {
		switch messageType {
		case key.CloseExternalPersistentConnectionMessageType:
			connectionID, _ := headers[key.ExternalConnectionIDKey]
			log.Printf("close connection: %s", connectionID)
			c.chanCloseExternalConnection <- connectionID
		}
	} else if externalConnectionID, ok := headers[key.ExternalConnectionIDKey]; ok {
		// Case when the message should be forward to the external connection
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
