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
	Connection   inter.ListenConnection
}

type ChanProxyMessageToInternal struct {
	Host    string
	Content []byte
}
