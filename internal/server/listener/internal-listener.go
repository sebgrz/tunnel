package listener

import (
	"fmt"
	"log"
	"net"
	"proxy/internal/server/connection"
	"proxy/internal/server/enum"
	"proxy/internal/server/inter"
	"proxy/internal/server/pack"
	"proxy/pkg/communication"
	"proxy/pkg/key"
	"sync"

	pkgenum "proxy/pkg/enum"
)

type InternalListener struct {
	port                                  string
	connections                           map[string]inter.ListenSendWithHeadersConnection
	chanAddConnection                     chan pack.ChanInternalConnection
	chanRemoveConnection                  chan pack.ChanInternalConnectionSpec
	chanCloseExternalConnection           chan<- string
	chanReceivedInternalMessageToExternal chan pack.ChanProxyMessageToExternal
}

func NewInternalListener(
	port string,
	chanMsgToInternal <-chan pack.ChanProxyMessageToInternal,
	chanMsgToExternal chan<- pack.ChanProxyMessageToExternal,
	chanCloseExternalConnection chan<- string,
) *InternalListener {
	l := &InternalListener{
		port:                                  port,
		connections:                           make(map[string]inter.ListenSendWithHeadersConnection),
		chanAddConnection:                     make(chan pack.ChanInternalConnection),
		chanRemoveConnection:                  make(chan pack.ChanInternalConnectionSpec),
		chanReceivedInternalMessageToExternal: make(chan pack.ChanProxyMessageToExternal),
		chanCloseExternalConnection:           chanCloseExternalConnection,
	}
	go func() {
		mu := sync.Mutex{}
		for {
			select {
			case msgToExternal := <-l.chanReceivedInternalMessageToExternal:
				chanMsgToExternal <- msgToExternal
			case sendMessage := <-chanMsgToInternal:
				mu.Lock()
				if con, ok := l.connections[mapKey(sendMessage.ConnectionType, sendMessage.Host)]; ok {
					switch sendMessage.MessageType {
					case enum.MessageExternalToInternalMessageType:
						err := con.Send(sendMessage.ExternalConnectionID, sendMessage.Content)
						if err != nil {
							log.Print(err)
						}
					case enum.CloseConnectionExternalToInternalMessageType:
						headers := communication.BytesHeader{
							key.MessageTypeBytesHeader: key.CloseExternalPersistentConnectionMessageType,
						}
						err := con.SendWithHeaders(sendMessage.ExternalConnectionID, headers, sendMessage.Content)
						if err != nil {
							log.Print(err)
						}
					}
				}
				mu.Unlock()
			case addConnection := <-l.chanAddConnection:
				l.connections[mapKey(addConnection.ConnectionType, addConnection.Host)] = addConnection.Connection
				log.Printf("internal connection: %s of type [%s] added", addConnection.Host, addConnection.ConnectionType)
			case removeConnection := <-l.chanRemoveConnection:
				mu.Lock()
				delete(l.connections, mapKey(removeConnection.ConnectionType, removeConnection.Host))
				log.Printf("internal connection: %s of type [%s] removed", removeConnection, "TYPE")
				mu.Unlock()
			}
			log.Printf("internal connections number: %d", len(l.connections))
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
			ec := connection.NewInternalConnection(con, l.chanRemoveConnection, l.chanCloseExternalConnection, l.chanAddConnection, l.chanReceivedInternalMessageToExternal)
			go ec.Listen()
		}
	}()
}

func mapKey(connectionType pkgenum.AgentConnectionType, hostname string) string {
	return fmt.Sprintf("%s_%s", connectionType, hostname)
}
