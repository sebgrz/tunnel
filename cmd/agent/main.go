package main

import (
	"fmt"
	"net"
)

func main() {
	msg := `{
  "id": "1",
	"corr_id": "1234",
  "type": "AgentRegistrationMessage",
  "hostname": "proxy.local"
}`

	con, _ := net.Dial("tcp", "192.168.1.13:5050")
	con.Write([]byte(msg))

	msgBytes := make([]byte, 0)
	for {
		b := make([]byte, 1024)
		bl, _ := con.Read(b)
		msgBytes = append(msgBytes, b[:bl]...)
		if bl < len(b) {
			fmt.Println(string(msgBytes))
			msgBytes = make([]byte, 0)
			continue
		}
	}
	con.Close()
}
