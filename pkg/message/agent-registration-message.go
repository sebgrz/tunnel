package message

import (
	"proxy/pkg/enum"

	goeh "github.com/hetacode/go-eh"
)

type AgentRegistrationMessage struct {
	*goeh.EventData
	Hostname string `json:"hostname"`
	ConnectionType enum.AgentConnectionType `json:"conn_type"`
}

func (e *AgentRegistrationMessage) GetType() string {
	return "AgentRegistrationMessage"
}
