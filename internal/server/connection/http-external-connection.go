package connection

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"net"
	"net/http"
	"proxy/internal/server/enum"
	"proxy/internal/server/inter"
	"proxy/internal/server/pack"
	pkgenum "proxy/pkg/enum"
	"proxy/pkg/helper"
	"strings"
	"time"

	"github.com/hashicorp/go-uuid"
)

type HTTPExternalConnection struct {
	ID                   string
	host                 string
	connectionType       pkgenum.AgentConnectionType
	connection           net.Conn
	initialData          *inter.ExternalConnectionInitialData
	chanRemoveConnection chan<- string
	chanMsgToInternal    chan<- pack.ChanProxyMessageToInternal
	chanIncomingMessage  chan []byte
}

func NewHTTPExternalConnection(con net.Conn, chanRemoveConnection chan<- string, chanMsgToInternal chan<- pack.ChanProxyMessageToInternal) *HTTPExternalConnection {
	id, _ := uuid.GenerateUUID()
	c := &HTTPExternalConnection{
		ID:                   id,
		connectionType:       pkgenum.HTTPAgentConnectionType,
		connection:           con,
		chanRemoveConnection: chanRemoveConnection,
		chanMsgToInternal:    chanMsgToInternal,
		chanIncomingMessage:  make(chan []byte),
	}
	return c
}

func (c *HTTPExternalConnection) Send(externalConnectionID string, msgBytes []byte) error {
	if c.connection == nil {
		return fmt.Errorf("external connection is not initialized")
	}
	c.chanIncomingMessage <- msgBytes

	return nil
}

func (c *HTTPExternalConnection) Listen() {
	log.Printf("New connection: %s", c.connection.RemoteAddr().String())

	var err error
	var msgBytes []byte
	if c.initialData != nil && c.initialData.MsgBytes != nil {
		msgBytes = *c.initialData.MsgBytes
		c.initialData.MsgBytes = nil
	} else {
		msgBytes, err = helper.RecvBytes(c.connection)
		if err != nil {
			log.Fatal(err)
		}
	}
	log.Printf("%s", string(msgBytes))

	var httpRequest *http.Request
	if c.initialData != nil && c.initialData.Request != nil {
		httpRequest = c.initialData.Request
		c.initialData.Request = nil
	} else {
		br := bufio.NewReader(bytes.NewReader(msgBytes))
		httpRequest, err = http.ReadRequest(br)
		if err != nil {
			c.chanRemoveConnection <- c.ID
			log.Fatal(err)
		}
	}

	hostArr := strings.Split(httpRequest.Host, ":") // <hostname>:<port>
	c.host = hostArr[0]
	log.Printf("HOST: %s", hostArr[0])
	c.chanMsgToInternal <- pack.ChanProxyMessageToInternal{
		ExternalConnectionID: c.ID,
		Host:                 hostArr[0],
		ConnectionType:       c.connectionType,
		MessageType:          enum.MessageExternalToInternalMessageType,
		Content:              msgBytes,
	}
	log.Printf("End receiving")

	// Response
	select {
	// Timeout
	case <-time.Tick(time.Second * 30):
		timeoutMessage := `HTTP/1.1 504 Gateway Timeout
		Server: HetaProxy
		Connection: Closed
		Content-Type: text/html; charset=utf-8`
		msgBytes = []byte(timeoutMessage)
		log.Printf("external connection: %s timeout", c.ID)
		// Message
	case msg := <-c.chanIncomingMessage:
		msgBytes = msg
		log.Printf("external connection: %s incoming message", c.ID)
	}
	_, err = c.connection.Write(msgBytes)
	if err != nil {
		c.chanRemoveConnection <- c.ID
		log.Fatal(err)
	}

	err = c.connection.Close()
	if err != nil {
		c.chanRemoveConnection <- c.ID
		log.Fatal(err)
	}
	c.chanRemoveConnection <- c.ID
}

func (c *HTTPExternalConnection) InitialData(data *inter.ExternalConnectionInitialData) {
	c.initialData = data
}

func (c *HTTPExternalConnection) GetID() string {
	return c.ID
}

func (c *HTTPExternalConnection) GetHost() string {
	return c.host
}

func (c *HTTPExternalConnection) Close() {
	if c.connection != nil {
		c.connection.Close()
	}
}

func (c *HTTPExternalConnection) GetConnectionType() pkgenum.AgentConnectionType {
	return c.connectionType
}
