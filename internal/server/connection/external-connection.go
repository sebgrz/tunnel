package connection

import (
	"bufio"
	"bytes"
	"io"
	"log"
	"net"
	"net/http"
	"proxy/internal/server/pack"
	"strings"

	"github.com/hashicorp/go-uuid"
)

type ExternalConnection struct {
	ID                   string
	connection           net.Conn
	chanRemoveConnection chan<- string
	chanMsgToInternal    chan<- pack.ChanProxyMessageToInternal
}

func NewExternalConnection(con net.Conn, chanRemoveConnection chan<- string, chanMsgToInternal chan<- pack.ChanProxyMessageToInternal) *ExternalConnection {
	id, _ := uuid.GenerateUUID()
	c := &ExternalConnection{
		ID:                   id,
		connection:           con,
		chanRemoveConnection: chanRemoveConnection,
		chanMsgToInternal:    chanMsgToInternal,
	}
	return c
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

	// TODO: wait foe agent response

	err = c.connection.Close()
	if err != nil {
		c.chanRemoveConnection <- c.ID
		log.Fatal(err)
	}
	c.chanRemoveConnection <- c.ID
}
