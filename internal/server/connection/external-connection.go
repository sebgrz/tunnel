package connection

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"proxy/internal/server/pack"
	"strings"
	"time"

	"github.com/hashicorp/go-uuid"
)

type ExternalConnection struct {
	ID                   string
	connection           net.Conn
	chanRemoveConnection chan<- string
	chanMsgToInternal    chan<- pack.ChanProxyMessageToInternal
	chanIncomingMessage  chan []byte
}

func NewExternalConnection(con net.Conn, chanRemoveConnection chan<- string, chanMsgToInternal chan<- pack.ChanProxyMessageToInternal) *ExternalConnection {
	id, _ := uuid.GenerateUUID()
	c := &ExternalConnection{
		ID:                   id,
		connection:           con,
		chanRemoveConnection: chanRemoveConnection,
		chanMsgToInternal:    chanMsgToInternal,
		chanIncomingMessage:  make(chan []byte),
	}
	return c
}

func (c *ExternalConnection) Send(externalConnectionID string, msgBytes []byte) error {
	if c.connection == nil {
		return fmt.Errorf("external connection is not initialized")
	}
	c.chanIncomingMessage <- msgBytes

	return nil
}

func (c *ExternalConnection) Listen() {
	log.Printf("New connection: %s", c.connection.RemoteAddr().String())

	msgBytes := make([]byte, 0)
	for {
		b := make([]byte, 1024)
		bl, err := c.connection.Read(b)
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Fatal(err)
		}
		log.Printf("Recv bytes: %d", bl)
		msgBytes = append(msgBytes, b[:bl]...)
		if bl < len(b) {
			break
		}
	}

	br := bufio.NewReader(bytes.NewReader(msgBytes))
	httpRequest, err := http.ReadRequest(br)
	if err != nil {
		c.chanRemoveConnection <- c.ID
		log.Fatal(err)
	}
	hostArr := strings.Split(httpRequest.Host, ":") // <hostname>:<port>
	log.Printf("HOST: %s", hostArr[0])
	c.chanMsgToInternal <- pack.ChanProxyMessageToInternal{
		ExternalConnectionID: c.ID,
		Host:                 hostArr[0],
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
