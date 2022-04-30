package agent

import (
	"log"
	"net"
	"proxy/internal/agent/client"
	"proxy/internal/agent/configuration"
	"proxy/internal/agent/http"
	"proxy/internal/agent/pack"
	"proxy/pkg/communication"
	"proxy/pkg/enum"
	"proxy/pkg/key"
	"proxy/pkg/message"
	"sync"
	"time"

	"github.com/hashicorp/go-uuid"
	goeh "github.com/hetacode/go-eh"
)

type CallDestinationHandler func(headers communication.BytesHeader, msg []byte)

type Agent struct {
	config                *configuration.Configuration
	waitingConnections    map[string]CallDestinationHandler
	persistentConnections map[string]*client.WSClient // TODO: maybe some abstraction - interface for different persistent connections
}

func NewAgent(config *configuration.Configuration, serverAddress, hostnameListener, destinationAddress string, connectionType enum.AgentConnectionType) *Agent {
	if config == nil {
		config = &configuration.Configuration{
			ServerAddress: serverAddress,
			Destination: configuration.Destination{
				Hostname:       hostnameListener,
				ConnectionType: connectionType,
				Proxy: []configuration.DestinationProxy{
					{
						OriginHostname:     hostnameListener,
						DestinationAddress: destinationAddress,
					},
				},
			},
		}
	}

	a := &Agent{
		config:                config,
		waitingConnections:    make(map[string]CallDestinationHandler),
		persistentConnections: make(map[string]*client.WSClient),
	}
	return a
}

func (a *Agent) Start() {
	uid, _ := uuid.GenerateUUID()

	con, err := net.Dial("tcp", a.config.ServerAddress)
	if err != nil {
		log.Fatalf("config %+v err: %s", a.config, err)
	}

	// Send agent registration message
	// That message register this specific agent on the server side
	time.Sleep(2 * time.Second)
	msgRegistrationBytes := communication.SerializeBytesMessage(nil, createAgentRegistrationMessage(uid, a.config.Destination.Hostname, a.config.Destination.ConnectionType))
	con.Write(msgRegistrationBytes)

	chanAddProxyConnection := make(chan []byte)
	chanRemoveProxyConnection := make(chan string)
	chanRemoveProxyPersistentConnection := make(chan string)
	chanSendResponse := make(chan pack.ChanResponseToServer)
	chanSendPersistentResponse := make(chan pack.ChanResponseToServer)

	go func() {
		responseMutex := sync.Mutex{}
		responsePersistentMutex := sync.Mutex{}
		removeProxyMutex := sync.Mutex{}
		for {
			select {
			case msg := <-chanAddProxyConnection:
				headers, msgBytes := communication.DeserializeBytesMessage(msg)
				if connectionID, ok := headers[key.ExternalConnectionIDKey]; ok {
					log.Printf("msg received for connection: %s", connectionID)

					var originHostname string
					if originHostname, ok = headers[key.OriginHostname]; !ok {
						log.Fatalf("connection_id %s - no OriginHostname header", connectionID)
					}

					destinationAddress, err := a.config.Destination.GetDestinationAddress(originHostname)
					if err != nil {
						log.Fatalf("connection_id %s - %s", connectionID, err)
					}

					handler := func(headers communication.BytesHeader, msg []byte) {
						switch a.config.Destination.ConnectionType {
						case enum.HTTPAgentConnectionType:
							// Set up timeout
							go func(connectionID string) {
								<-time.Tick(30 * time.Second)
								chanRemoveProxyConnection <- connectionID
							}(connectionID)

							// 1. Call destination address and wait for response
							response, err := http.Send(destinationAddress, msg)
							if err != nil {
								log.Println(err)
								return
							}
							// 3. Combine headers and response into bytes message
							responseMsg := communication.SerializeBytesMessage(headers, response)
							// 4. Send bytes message to the SendResponse channel
							chanSendResponse <- pack.ChanResponseToServer{
								ConnectionID:    connectionID,
								ResponseMessage: responseMsg,
							}
						case enum.WSAgentConnectionType:
							var persistentConnection *client.WSClient
							if persistentConnection, ok = a.persistentConnections[connectionID]; !ok {
								// 1. Create websocket connection to destinationAddress
								persistentConnection = client.NewWSClient(connectionID, destinationAddress, chanSendPersistentResponse, chanRemoveProxyPersistentConnection)
								// 2. Save connection in map
								a.persistentConnections[connectionID] = persistentConnection
								go persistentConnection.Listen()
							}

							if msgType, ok := headers[key.MessageTypeBytesHeader]; ok {
								switch msgType {
								case key.CloseExternalPersistentConnectionMessageType:
									if conn, ok := a.persistentConnections[connectionID]; ok {
										conn.Close()
									}
								}
							} else {
								persistentConnection.Send(msgBytes)
							}
						}
					}
					a.waitingConnections[connectionID] = handler
					go handler(headers, msgBytes)
				}
			case connectionID := <-chanRemoveProxyPersistentConnection:
				removeProxyMutex.Lock()
				log.Printf("remove persistent connection: %s", connectionID)
				headers := communication.BytesHeader{
					key.ExternalConnectionIDKey: connectionID,
					key.MessageTypeBytesHeader:  key.CloseExternalPersistentConnectionMessageType,
				}
				respMsgBytes := communication.SerializeBytesMessage(headers, []byte(""))
				con.Write(respMsgBytes)
				delete(a.waitingConnections, connectionID)
				removeProxyMutex.Unlock()
			case connectionID := <-chanRemoveProxyConnection:
				removeProxyMutex.Lock()
				log.Printf("remove connection: %s", connectionID)
				delete(a.waitingConnections, connectionID)
				removeProxyMutex.Unlock()
			case response := <-chanSendResponse:
				responseMutex.Lock()
				con.Write(response.ResponseMessage)
				go func() { chanRemoveProxyConnection <- response.ConnectionID }()
				responseMutex.Unlock()
			case response := <-chanSendPersistentResponse:
				responsePersistentMutex.Lock()
				headers := communication.BytesHeader{
					key.ExternalConnectionIDKey: response.ConnectionID,
				}
				respMsgBytes := communication.SerializeBytesMessage(headers, response.ResponseMessage)
				con.Write(respMsgBytes)
				responsePersistentMutex.Unlock()
			}
		}
	}()

	msgBytes := make([]byte, 0)
	for {
		b := make([]byte, 1024)
		bl, err := con.Read(b)
		if err != nil {
			log.Printf("err: %s", err)
			break
		}
		msgBytes = append(msgBytes, b[:bl]...)
		if bl < len(b) {
			chanAddProxyConnection <- msgBytes

			msgBytes = make([]byte, 0)
			continue
		}
	}
	con.Close()
}

func createAgentRegistrationMessage(uid string, hostnameListener string, connectionType enum.AgentConnectionType) []byte {
	msg := &message.AgentRegistrationMessage{
		EventData: &goeh.EventData{
			ID:            uid,
			CorrelationID: uid,
		},
		Hostname:       hostnameListener,
		ConnectionType: connectionType,
	}
	msg.SavePayload(msg)
	msgBytes := []byte(msg.GetPayload())

	return msgBytes
}
