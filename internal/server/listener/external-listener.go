package listener

import (
	"fmt"
	"log"
	"net"
	"proxy/internal/server/connection"
)

type ChanConnection struct {
	connectionID string
	connection   *connection.ExternalConnection
}

type ExternalListener struct {
	port                 string
	connections          map[string]*connection.ExternalConnection
	chanAddConnection    chan ChanConnection
	chanRemoveConnection chan string
}

func NewExternalListener(port string) *ExternalListener {
	l := &ExternalListener{
		port:                 port,
		connections:          make(map[string]*connection.ExternalConnection),
		chanAddConnection:    make(chan ChanConnection),
		chanRemoveConnection: make(chan string),
	}
	go func() {
		for {
			select {
			case addConnection := <-l.chanAddConnection:
				l.connections[addConnection.connectionID] = addConnection.connection
				log.Printf("connection: %s added", addConnection.connectionID)
			case removeConnection := <-l.chanRemoveConnection:
				delete(l.connections, removeConnection)
				log.Printf("connection: %s removed", removeConnection)
			}
			log.Printf("connections: %d", len(l.connections))
		}
	}()
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
			ec := connection.NewExternalConnection(con, l.chanRemoveConnection)
			l.chanAddConnection <- ChanConnection{
				connectionID: ec.ID,
				connection:   ec,
			}
			go ec.Listen()
		}
	}()
}
