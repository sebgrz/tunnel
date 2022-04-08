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
	proxynet "proxy/pkg/net"
	"strings"

	"github.com/gorilla/websocket"
	"github.com/hashicorp/go-uuid"
)

type WSExternalConnection struct {
	ID                   string
	connectionType       pkgenum.AgentConnectionType
	host                 string
	connection           net.Conn
	wsConnection         *websocket.Conn
	initialData          *inter.ExternalConnectionInitialData
	chanRemoveConnection chan<- string
	chanMsgToInternal    chan<- pack.ChanProxyMessageToInternal
}

func NewWSExternalConnection(con net.Conn, chanRemoveConnection chan<- string, chanMsgToInternal chan<- pack.ChanProxyMessageToInternal) *WSExternalConnection {
	id, _ := uuid.GenerateUUID()
	c := &WSExternalConnection{
		ID:                   id,
		connectionType:       pkgenum.WSAgentConnectionType,
		connection:           con,
		chanRemoveConnection: chanRemoveConnection,
		chanMsgToInternal:    chanMsgToInternal,
	}
	return c
}

func (c *WSExternalConnection) Send(externalConnectionID string, msgBytes []byte) error {
	if c.wsConnection == nil {
		return fmt.Errorf("external ws connection is not initialized")
	}
	err := c.wsConnection.WriteMessage(websocket.TextMessage, msgBytes)
	if err != nil {
		return err
	}
	return nil
}

func (c *WSExternalConnection) Listen() {
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

	var upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
	rww := proxynet.NewResponseWriterWrapper(c.connection, httpRequest)
	upgrader.CheckOrigin = func(r *http.Request) bool {
		log.Print("origin: " + r.RemoteAddr)
		return true
	}
	wsConn, err := upgrader.Upgrade(rww, httpRequest, nil)
	if err != nil {
		c.chanRemoveConnection <- c.ID
		log.Fatal(err)
	}
	defer wsConn.Close()

	c.wsConnection = wsConn
	hostArr := strings.Split(httpRequest.Host, ":")
	c.host = hostArr[0]

	for {
		_, msg, err := wsConn.ReadMessage()
		if err != nil {
			c.chanRemoveConnection <- c.ID
			log.Println(err)
			break
		}

		c.chanMsgToInternal <- pack.ChanProxyMessageToInternal{
			ExternalConnectionID: c.ID,
			Host:                 hostArr[0],
			ConnectionType:       c.connectionType,
			MessageType:          enum.MessageExternalToInternalMessageType,
			Content:              msg,
		}
	}

	c.chanRemoveConnection <- c.ID
}

func (c *WSExternalConnection) InitialData(data *inter.ExternalConnectionInitialData) {
	c.initialData = data
}

func (c *WSExternalConnection) GetID() string {
	return c.ID
}

func (c *WSExternalConnection) GetHost() string {
	return c.host
}

func (c *WSExternalConnection) Close() {
	if c.connection != nil {
		log.Printf("WSExternalConnection close: %s", c.ID)
		c.connection.Close()
	}
}

func (c *WSExternalConnection) GetConnectionType() pkgenum.AgentConnectionType {
	return c.connectionType
}
