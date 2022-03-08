package listener

import (
	"fmt"
	"log"
	"net"
	"proxy/internal/server/connection"
	"proxy/internal/server/inter"
	"proxy/internal/server/pack"
	"sync"
)

type ExternalListener struct {
	port                 string
	connections          map[string]inter.ListenSendConnection
	chanAddConnection    chan pack.ChanExternalConnection
	chanRemoveConnection chan string
	chanMsgToInternal    chan<- pack.ChanProxyMessageToInternal
	chanMsgToExternal    <-chan pack.ChanProxyMessageToExternal
}

func NewExternalListener(port string, chanMsgToInternal chan<- pack.ChanProxyMessageToInternal, chanMsgToExternal <-chan pack.ChanProxyMessageToExternal) *ExternalListener {
	l := &ExternalListener{
		port:                 port,
		connections:          make(map[string]inter.ListenSendConnection),
		chanAddConnection:    make(chan pack.ChanExternalConnection),
		chanRemoveConnection: make(chan string),
		chanMsgToInternal:    chanMsgToInternal,
		chanMsgToExternal:    chanMsgToExternal,
	}

	go func() {
		mu := sync.Mutex{}
		for {
			select {
			case msgToExternal := <-chanMsgToExternal:
				mu.Lock()
				if con, ok := l.connections[msgToExternal.ExternalConnectionID]; ok {
					err := con.Send(msgToExternal.ExternalConnectionID, msgToExternal.Content)
					if err != nil {
						log.Print(err)
					}
				}
				mu.Unlock()
			case addConnection := <-l.chanAddConnection:
				l.connections[addConnection.ConnectionID] = addConnection.Connection
				log.Printf("external connection: %s added", addConnection.ConnectionID)
			case removeConnection := <-l.chanRemoveConnection:
				mu.Lock()
				delete(l.connections, removeConnection)
				mu.Unlock()
				log.Printf("external connection: %s removed", removeConnection)
			}
			log.Printf("external connections: %d", len(l.connections))
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
