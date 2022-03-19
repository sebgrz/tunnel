package agent

import (
	"log"
	"net"
	"proxy/internal/agent/http"
	"proxy/internal/agent/pack"
	"proxy/pkg/communication"
	"proxy/pkg/key"
	"proxy/pkg/message"
	"sync"
	"time"

	"github.com/hashicorp/go-uuid"
	goeh "github.com/hetacode/go-eh"
)

type CallDestinationHandler func(headers communication.BytesHeader, msg []byte)

type Agent struct {
	serverAddress      string
	hostnameListener   string
	destinationAddress string

	waitingConnections map[string]CallDestinationHandler
}

func NewAgent(serverAddress, hostnameListener, destinationAddress string) *Agent {
	a := &Agent{
		serverAddress:      serverAddress,
		hostnameListener:   hostnameListener,
		destinationAddress: destinationAddress,
		waitingConnections: make(map[string]CallDestinationHandler),
	}
	return a
}

func (a *Agent) Start() {
	uid, _ := uuid.GenerateUUID()

	con, err := net.Dial("tcp", a.serverAddress)
	if err != nil {
		log.Fatal(err)
	}

	// Send agent registration message
	// That message register this specific agent on the server side
	time.Sleep(2 * time.Second)
	msgRegistrationBytes := communication.SerializeBytesMessage(nil, createAgentRegistrationMessage(uid, a.hostnameListener))
	con.Write(msgRegistrationBytes)

	chanAddProxyConnection := make(chan []byte)
	chanRemoveProxyConnection := make(chan string)
	chanSendResponse := make(chan pack.ChanResponseToServer)

	go func() {
		responseMutex := sync.Mutex{}
		removeProxyMutex := sync.Mutex{}
		for {
			select {
			case msg := <-chanAddProxyConnection:
				headers, msgBytes := communication.DeserializeBytesMessage(msg)
				if connectionID, ok := headers[key.ExternalConnectionIDKey]; ok {
					log.Printf("msg received for connection: %s", connectionID)
					handler := func(headers communication.BytesHeader, msg []byte) {
						// Set up timeout
						go func(connectionID string) {
							<-time.Tick(30 * time.Second)
							chanRemoveProxyConnection <- connectionID
						}(connectionID)
						// 1. Call destination address and wait for response
						response, err := http.Send(a.destinationAddress, msg)
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
					}
					a.waitingConnections[connectionID] = handler
					go handler(headers, msgBytes)
				}
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

func createAgentRegistrationMessage(uid string, hostnameListener string) []byte {
	msg := &message.AgentRegistrationMessage{
		EventData: &goeh.EventData{
			ID:            uid,
			CorrelationID: uid,
		},
		Hostname: hostnameListener,
	}
	msg.SavePayload(msg)
	msgBytes := []byte(msg.GetPayload())

	return msgBytes
}
