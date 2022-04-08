package server

import (
	"proxy/internal/server/configuration"
	"proxy/internal/server/listener"
	"proxy/internal/server/pack"
)

type Server struct {
	extListener       *listener.ExternalListener
	intListener       *listener.InternalListener
	chanMsgToInternal chan pack.ChanProxyMessageToInternal
	chanMsgToExternal chan pack.ChanProxyMessageToExternal
}

func NewServer(config *configuration.Configuration, externalPort, externalSslPort, advertisingAgentPort string) *Server {
	// request message client -> <request> -> external connection -> external listener -> internal listener -> internal connection -> agent -> web server
	chanMsgToInternal := make(chan pack.ChanProxyMessageToInternal)
	// response message -> web server -> agent -> internal connection -> internal listener -> external listener -> external connection -> client
	chanMsgToExternal := make(chan pack.ChanProxyMessageToExternal)
	// inform external connection that agent <-> destination connection is closed (persistent case)
	chanAgentConnectionClosedToExternal := make(chan string)
	// inform agent <-> destination connection that external connection is closed (persistent case)
	chanExternalConnectionClosedToAgent := make(chan string)
	s := &Server{
		chanMsgToInternal: chanMsgToInternal,
		intListener:       listener.NewInternalListener(advertisingAgentPort, chanMsgToInternal, chanMsgToExternal, chanAgentConnectionClosedToExternal),
		extListener:       listener.NewExternalListener(config, externalPort, externalSslPort, chanMsgToInternal, chanMsgToExternal, chanAgentConnectionClosedToExternal, chanExternalConnectionClosedToAgent),
	}
	return s
}

func (s *Server) Start() {
	s.extListener.Run()
	s.intListener.Run()
}
