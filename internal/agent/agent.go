package agent

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"proxy/internal/agent/http"
	"proxy/pkg/communication"
	"proxy/pkg/message"
	"time"

	"github.com/hashicorp/go-uuid"
	goeh "github.com/hetacode/go-eh"
)

type Agent struct {
	serverAddress      string
	hostnameListener   string
	destinationAddress string
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
			fmt.Println(string(msgBytes))
			// TODO: headers
			_, msgBytes := communication.DeserializeBytesMessage(msgBytes)

			msgBytes = make([]byte, 0)
			// TODO:
			// 2. save connection in memory
			// 3. send message in goroutine
			// 4. receive response and forward to the connection via channel?
			http.Send(a.destinationAddress, msgBytes)
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
