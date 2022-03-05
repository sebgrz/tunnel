package main

import "net"

func main() {
	msg := `{
  "id": "1",
	"corr_id": "1234",
  "type": "AgentRegistrationMessage",
  "hostname": "proxy.local"
}`

	con, _ := net.Dial("tcp", "192.168.1.13:5050")
	con.Write([]byte(msg))
}
