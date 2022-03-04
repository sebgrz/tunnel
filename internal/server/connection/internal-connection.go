package connection

import (
	"fmt"
	"io"
	"log"
	"net"

	"github.com/hashicorp/go-uuid"
)

type InternalConnection struct {
	ID                   string
	Host                 string
	connection           net.Conn
	chanRemoveConnection chan<- string
}

func NewInternalConnection(con net.Conn, chanRemoveConnection chan<- string) *InternalConnection {
	id, _ := uuid.GenerateUUID()
	c := &InternalConnection{
		ID:                   id,
		connection:           con,
		chanRemoveConnection: chanRemoveConnection,
	}
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
				parseBytesMessage(msgBytes)
				msgBytes = make([]byte, 0)
				continue
			}
			c.chanRemoveConnection <- c.ID
			log.Print(err)
			break
		}
		log.Printf("Recv bytes: %d", bl)
		msgBytes = append(msgBytes, b...)
		if bl == 0 {
			parseBytesMessage(msgBytes)
			c.chanRemoveConnection <- c.ID
			continue
		}
	}
	log.Printf("End receiving")
	c.chanRemoveConnection <- c.ID
}

func parseBytesMessage(msgBytes []byte) {
	// TODO: message handler
}
