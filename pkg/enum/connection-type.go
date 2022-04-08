package enum

type AgentConnectionType string

const (
	HTTPAgentConnectionType AgentConnectionType = "http"
	WSAgentConnectionType   AgentConnectionType = "ws"
)
