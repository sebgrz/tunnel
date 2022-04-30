package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"proxy/internal/agent"
	"proxy/internal/agent/configuration"
	"proxy/pkg/enum"
	"proxy/pkg/helper"
)

var (
	serverAddress      *string
	hostnameListener   *string
	destinationAddress *string
)

func main() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Printf("get home dir err: %s", err)
	}

	serverAddress = flag.String("server", "localhost:5050", "Server address")
	hostnameListener = flag.String("hostname", "proxy.local", "Hostname use as recognizer of given flow")
	destinationAddress = flag.String("destination", "http://localhost:4321", "Address of where exists web application")
	connectionType := flag.String("type", "http", "Choose connection type: http|ws")
	configPath := flag.String("config", fmt.Sprintf("%s/.config/tunnel/agent.json", homeDir), "Configuration file")
	flag.Parse()

	config, err := helper.LoadJsonFile[configuration.Configuration](*configPath)

	agent := agent.NewAgent(config, *serverAddress, *hostnameListener, *destinationAddress, enum.AgentConnectionType(*connectionType))
	log.Println("agent is starting")
	agent.Start()

	log.Println("agent is stopped")
}
