package pack

import (
	"proxy/internal/server/enum"
	"proxy/internal/server/inter"
)

type ChanInternalConnection struct {
	Host       string
	Connection inter.ListenSendWithHeadersConnection
}

type ChanExternalConnection struct {
	ConnectionID string
	Connection   inter.ExternalConnection
}

type ChanProxyMessageToInternal struct {
	ExternalConnectionID string
	Host                 string
	Type                 enum.ExternalToInternalMessageType
	Content              []byte
}

type ChanProxyMessageToExternal struct {
	ExternalConnectionID string
	Content              []byte
}
