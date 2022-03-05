package message

import goeh "github.com/hetacode/go-eh"

type AgentRegistrationMessage struct {
	*goeh.EventData
	Hostname string `json:"hostname"`
}

func (e *AgentRegistrationMessage) GetType() string {
	return "AgentRegistrationMessage"
}
