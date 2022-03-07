package listener

import (
	"fmt"
	"log"
	"net"
	"proxy/internal/server/connection"
	"proxy/internal/server/inter"
	"proxy/internal/server/pack"
)

type ExternalListener struct {
	port                 string
	connections          map[string]inter.ListenConnection
	chanAddConnection    chan pack.ChanExternalConnection
	chanRemoveConnection chan string
	chanMsgToInternal    chan<- pack.ChanProxyMessageToInternal
	chanMsgToExternal    <-chan pack.ChanProxyMessageToExternal
}

func NewExternalListener(port string, chanMsgToInternal chan<- pack.ChanProxyMessageToInternal) *ExternalListener {
	l := &ExternalListener{
		port:                 port,
		connections:          make(map[string]inter.ListenConnection),
		chanAddConnection:    make(chan pack.ChanExternalConnection),
		chanRemoveConnection: make(chan string),
		chanMsgToInternal:    chanMsgToInternal,
	}

	go func() {
		for {
			select {
			case addConnection := <-l.chanAddConnection:
				l.connections[addConnection.ConnectionID] = addConnection.Connection
				log.Printf("connection: %s added", addConnection.ConnectionID)
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
			ec := connection.NewExternalConnection(con, l.chanRemoveConnection, l.chanMsgToInternal)
			l.chanAddConnection <- pack.ChanExternalConnection{
				ConnectionID: ec.ID,
				Connection:   ec,
			}
			go ec.Listen()
		}
	}()
}
