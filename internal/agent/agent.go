package agent

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"proxy/internal/agent/http"
	"proxy/internal/agent/pack"
	"proxy/pkg/communication"
	"proxy/pkg/key"
	"proxy/pkg/message"
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
	con.Write(createAgentRegistrationMessage(uid, a.hostnameListener))

	chanAddProxyConnection := make(chan []byte)
	chanRemoveProxyConnection := make(chan string)
	chanSendResponse := make(chan pack.ChanResponseToServer)

	go func() {
		for {
			select {
			case msg := <-chanAddProxyConnection:
				headers, msgBytes := communication.DeserializeBytesMessage(msg)
				if connectionID, ok := headers[key.ExternalConnectionIDKey]; ok {
					handler := func(headers communication.BytesHeader, msg []byte) {

						// 1. Call destination address and wait for response
						http.Send(a.destinationAddress, msg)

						// TODO
						// 3. Combine headers and response into bytes message
						// 4. Send bytes message to the SendResponse channel
					}
					a.waitingConnections[connectionID] = handler
					handler(headers, msgBytes)
				}
			case connectionID := <-chanRemoveProxyConnection:
				delete(a.waitingConnections, connectionID)
			case response := <-chanSendResponse:
				// TODO mutexes
				con.Write(response.ResponseMessage)
				delete(a.waitingConnections, response.ConnectionID)
			}
		}
	}()

	msgBytes := make([]byte, 0)
	for {
		b := make([]byte, 1024)
		bl, err := con.Read(b)
		if err != nil {
			fmt.Printf("err: %s", err)
			break
		}
		msgBytes = append(msgBytes, b[:bl]...)
		if bl < len(b) {
			// TODO: headers
			chanAddProxyConnection <- msgBytes

			msgBytes = make([]byte, 0)
			// TODO:
			// 3. send message in goroutine
			// 4. receive response and forward to the connection via channel?
			continue
		}
	}
	con.Close()
}

func createAgentRegistrationMessage(uid string, hostnameListener string) []byte {
	msg := message.AgentRegistrationMessage{
		EventData: &goeh.EventData{
			ID:            uid,
			CorrelationID: uid,
		},
		Hostname: hostnameListener,
	}
	msgBytes, _ := json.Marshal(msg)

	return msgBytes
}
