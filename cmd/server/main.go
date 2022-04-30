package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"proxy/internal/server"
	"proxy/internal/server/configuration"
	"proxy/pkg/helper"
)

var (
	externalPort         *string
	baseHostname         *string
	advertisingAgentPort *string
)

func main() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Printf("get home dir err: %s", err)
	}

	externalPort = flag.String("port", "5000", "External port")
	externalSslPort := flag.String("port-ssl", "", "External port for encrypted connection")
	advertisingAgentPort = flag.String("advPort", "5050", "Advertising agent port")
	baseHostname = flag.String("hostname", "", "Base hostname - it will be use to generate subdomains")
	configPath := flag.String("config", fmt.Sprintf("%s/.config/tunnel/server.json", homeDir), "Configuration file")
	flag.Parse()
	
	// Load configuration
	config, err := helper.LoadJsonFile[configuration.Configuration](*configPath)

	server := server.NewServer(config, *externalPort, *externalSslPort, *advertisingAgentPort)

	done := make(chan os.Signal)
	server.Start()
	<-done

	log.Print("server stopped")

}
