package listener

import (
	"fmt"
	"io"
	"log"
	"net"
)

type ExternalListener struct {
	port string
}

func NewExternalListener(port string) *ExternalListener {
	l := &ExternalListener{port}
	return l
}

func (l *ExternalListener) Run() {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%s", l.port))
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		for {
			con, err := listener.Accept()
			if err != nil {
				log.Fatal(err)
			}
			log.Printf("New connection: %s", con.RemoteAddr().String())

			for {
				b := make([]byte, 1024)
				bl, err := con.Read(b)
				if err != nil {
					if err == io.EOF {
						break
					}
					log.Fatal(err)
				}
				log.Printf("Recv bytes: %d", bl)
				if bl == 0 {
					break
				}
				log.Println(string(b))
				if bl < len(b) {
					break
				}

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

OK!
			`
			con.Write([]byte(responseMessage))
			con.Close()
		}
	}()
}
