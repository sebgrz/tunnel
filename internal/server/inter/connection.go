package inter

import (
	"net/http"
	"proxy/pkg/communication"
	"proxy/pkg/enum"
)

type ExternalConnectionInitialData struct {
	MsgBytes *[]byte
	Request  *http.Request
}

type SendConnection interface {
	Send(externalConnectionID, originHostname string, msgBytes []byte) error
}

type SendWithHeadersConnection interface {
	SendWithHeaders(externalConnectionID, originHostname string, headers communication.BytesHeader, msgBytes []byte) error
}

type ListenConnection interface {
	Listen()
}

type ConfigurableConnection interface {
	Configure(hostname string, connectionType enum.AgentConnectionType)
}

type ListenSendConnection interface {
	SendConnection
	ListenConnection
}

type ListenSendWithHeadersConnection interface {
	SendConnection
	SendWithHeadersConnection
	ListenConnection
}

type ExternalConnection interface {
	ListenSendConnection
	GetHost
	InitialData(data *ExternalConnectionInitialData)
	GetID() string
	GetConnectionType() enum.AgentConnectionType
	Close()
}
