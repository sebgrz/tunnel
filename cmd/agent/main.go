package main

import (
	"flag"
	"fmt"
	"net"
	"proxy/internal/agent/http"

	"github.com/hashicorp/go-uuid"
)

var (
	serverAddress      *string
	hostnameListener   *string
	destinationAddress *string
)

func main() {
	serverAddress = flag.String("server", "localhost:5050", "Server address")
	hostnameListener = flag.String("hostname", "proxy.local", "Hostname use as recognizer of given flow")
	destinationAddress = flag.String("destination", "localhost:4321", "Address of where exists web application")

	uid, _ := uuid.GenerateUUID()

	msg := fmt.Sprintf(`{
  "id": "%s",
	"corr_id": "%s",
  "type": "AgentRegistrationMessage",
  "hostname": "%s"
}`, uid, uid, *hostnameListener)

	con, _ := net.Dial("tcp", *serverAddress)
	con.Write([]byte(msg))

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
			msgBytes = make([]byte, 0)
			// TODO:
			// 1. deserialize message as BytesMessage
			// 2. save connection in memory
			// 3. send message in goroutine
			// 4. receive response and forward to the connection via channel?
			http.Send(*destinationAddress, msgBytes)
			continue
		}
	}
	con.Close()
}
