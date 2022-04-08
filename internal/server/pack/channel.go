package pack

import (
	"proxy/internal/server/enum"
	"proxy/internal/server/inter"
	pkgenum "proxy/pkg/enum"
)

type ChanInternalConnectionSpec struct {
	Host           string
	ConnectionType pkgenum.AgentConnectionType
}

type ChanInternalConnection struct {
	Host           string
	ConnectionType pkgenum.AgentConnectionType
	Connection     inter.ListenSendWithHeadersConnection
}

type ChanExternalConnection struct {
	ConnectionID string
	Connection   inter.ExternalConnection
}

type ChanProxyMessageToInternal struct {
	ExternalConnectionID string
	Host                 string
	ConnectionType       pkgenum.AgentConnectionType
	MessageType          enum.ExternalToInternalMessageType
	Content              []byte
}

type ChanProxyMessageToExternal struct {
	ExternalConnectionID string
	Content              []byte
}
