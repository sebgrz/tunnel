package messagehandler

import (
	"proxy/internal/server/inter"
	"proxy/pkg/message"

	goeh "github.com/hetacode/go-eh"
)

type AgentRegistrationMessageHandler struct {
	Connection inter.SetHost
}

// Execute message
func (h *AgentRegistrationMessageHandler) Handle(event goeh.Event) {
	e := event.(*message.AgentRegistrationMessage)
	h.Connection.SetHost(e.Hostname)
}
