package main

import (
	"flag"
	"log"
	"os"
	"proxy/internal/server"
)

var (
	externalPort         *string
	baseHostname         *string
	advertisingAgentPort *string
)

func main() {
	// TODO:
	// 1. external listener
	// 2, internal agents listener
	externalPort = flag.String("port", "5000", "External port")
	advertisingAgentPort = flag.String("advPort", "5050", "Advertising agent port")
	baseHostname = flag.String("hostname", "", "Base hostname - it will be use to generate subdomains")
	flag.Parse()

	server := server.NewServer(*externalPort, *advertisingAgentPort)

	done := make(chan os.Signal)
	server.Start()
	<-done

	log.Print("server stopped")

}
