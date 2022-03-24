package client

import (
	"log"
	"proxy/internal/agent/pack"

	"github.com/gorilla/websocket"
)

type WSClient struct {
	connectionID       string
	destinationAddress string
	connection         *websocket.Conn

	chanResponseMessage  chan<- pack.ChanResponseToServer
	chanConnectionClosed chan<- string
}

func NewWSClient(connectionID, destinationAddress string, chanResponseMessage chan<- pack.ChanResponseToServer, chanConnectionClosed chan<- string) *WSClient {
	c := &WSClient{
		connectionID:         connectionID,
		destinationAddress:   destinationAddress,
		chanResponseMessage:  chanResponseMessage,
		chanConnectionClosed: chanConnectionClosed,
	}

	return c
}

func (c *WSClient) Listen() {
	wsConn, _, err := websocket.DefaultDialer.Dial(c.destinationAddress, nil)
	if err != nil {
		log.Fatal(err)
	}
	c.connection = wsConn

	defer func() {
		log.Printf("ws connection %s closed", c.connectionID)
		wsConn.Close()
		c.chanConnectionClosed <- c.connectionID
	}()

	for {
		_, msg, err := wsConn.ReadMessage()
		if err != nil {
			log.Print(err)
			break
		}
		c.chanResponseMessage <- pack.ChanResponseToServer{
			ConnectionID:    c.connectionID,
			ResponseMessage: msg,
		}
	}
}

func (c *WSClient) Send(msgBytes []byte) error {
	if c.connection != nil {
		return c.connection.WriteMessage(websocket.TextMessage, msgBytes)
	}
	return nil
}

func (c *WSClient) Close() {
	c.connection.Close()
}
