package messagehandler

import (
	"proxy/internal/server/inter"
	"proxy/pkg/message"

	goeh "github.com/hetacode/go-eh"
)

type AgentRegistrationMessageHandler struct {
	Connection inter.ConfigurableConnection
}

// Execute message
func (h *AgentRegistrationMessageHandler) Handle(event goeh.Event) {
	e := event.(*message.AgentRegistrationMessage)
	h.Connection.Configure(e.Hostname, e.ConnectionType)
}
