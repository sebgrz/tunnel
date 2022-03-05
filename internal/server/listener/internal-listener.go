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

type InternalListener struct {
	port                 string
	connections          map[string]inter.ListenSendConnection
	chanAddConnection    chan pack.ChanInternalConnection
	chanRemoveConnection chan string
}

func NewInternalListener(port string, chanMsg <-chan pack.ChanProxyMessageToInternal) *InternalListener {
	l := &InternalListener{
		port:                 port,
		connections:          make(map[string]inter.ListenSendConnection),
		chanAddConnection:    make(chan pack.ChanInternalConnection),
		chanRemoveConnection: make(chan string),
	}
	go func() {
		mu := sync.Mutex{}
		for {
			select {
			case sendMessage := <-chanMsg:
				mu.Lock()
				if con, ok := l.connections[sendMessage.Host]; ok {
					err := con.Send(sendMessage.Content)
					if err != nil {
						log.Print(err)
					}
					mu.Unlock()
				}
			case addConnection := <-l.chanAddConnection:
				l.connections[addConnection.Host] = addConnection.Connection
				log.Printf("connection: %s added", addConnection.Host)
			case removeConnection := <-l.chanRemoveConnection:
				mu.Lock()
				delete(l.connections, removeConnection)
				log.Printf("connection: %s removed", removeConnection)
				mu.Unlock()
			}
			log.Printf("connections: %d", len(l.connections))
		}
	}()
	return l
}

func (l *InternalListener) Run() {
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
			ec := connection.NewInternalConnection(con, l.chanRemoveConnection, l.chanAddConnection)
			go ec.Listen()
		}
	}()
}
