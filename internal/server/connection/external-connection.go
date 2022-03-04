package connection

import (
	"bufio"
	"bytes"
	"io"
	"log"
	"net"
	"net/http"
)

type ExternalConnection struct {
	connection net.Conn
}

func NewExternalConnection(con net.Conn) *ExternalConnection {
	c := &ExternalConnection{con}
	return c
}

func (c *ExternalConnection) Listen() {
	log.Printf("New connection: %s", c.connection.RemoteAddr().String())

	b := make([]byte, 1024)
	bl, err := c.connection.Read(b)
	if err != nil {
		if err == io.EOF {
			return
		}
		log.Fatal(err)
	}
	log.Printf("Recv bytes: %d", bl)
	if bl == 0 {
		return
	}

	br := bufio.NewReader(bytes.NewReader(b))
	httpRequest, err := http.ReadRequest(br)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("HOST: %s", httpRequest.Host)
	for k, v := range httpRequest.Header {
		log.Printf("%s - %s", k, v)
	}

	log.Printf("End receiving")
	responseMessage := `
HTTP/1.1 200 OK
Date: Sun, 10 Oct 2010 23:26:07 GMT
Server: Proxy Server 
Last-Modified: Sun, 26 Sep 2010 22:04:35 GMT
ETag: "45b6-834-49130cc1182c0"
Accept-Ranges: bytes
Content-Length: 12
Connection: close
Content-Type: text/html

OK!`
	_, err = c.connection.Write([]byte(responseMessage))
	if err != nil {
		log.Fatal(err)
	}
	err = c.connection.Close()
	if err != nil {
		log.Fatal(err)
	}
}
