package configuration

import (
	"fmt"
	"proxy/pkg/enum"
)

type Configuration struct {
	ServerAddress string      `json:"server"`
	Destination   Destination `json:"destination"` // TODO: future: multi-destinations
}

type Destination struct {
	// Hostname: proxy.local or *.proxy.local
	Hostname       string                   `json:"hostname"`
	ConnectionType enum.AgentConnectionType `json:"connection_type"`
	Proxy          []DestinationProxy       `json:"proxy"`
}

type DestinationProxy struct {
	// OriginHostname: proxy.local or blog.proxy.local
	OriginHostname     string `json:"origin_hostname"`
	DestinationAddress string `json:"destination"`
}

func (d Destination) GetDestinationAddress(originHostname string) (string, error) {
	for _, proxy := range d.Proxy {
		if proxy.OriginHostname == originHostname {
			return proxy.DestinationAddress, nil
		}
	}
	return "", fmt.Errorf("originHostname %s cannot be found", originHostname)
}
