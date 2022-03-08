package pack

import (
	"proxy/internal/server/inter"
)

type ChanInternalConnection struct {
	Host       string
	Connection inter.ListenSendConnection
}

type ChanExternalConnection struct {
	ConnectionID string
	Connection   inter.ListenSendConnection
}

type ChanProxyMessageToInternal struct {
	ExternalConnectionID string
	Host                 string
	Content              []byte
}

type ChanProxyMessageToExternal struct {
	ExternalConnectionID string
	Content              []byte
}
