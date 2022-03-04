package listener

import (
	"fmt"
	"log"
	"net"
	"proxy/internal/server/connection"
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
			ec := connection.NewExternalConnection(con)
			go ec.Listen()
		}
	}()
}
