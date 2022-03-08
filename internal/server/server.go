package server

import (
	"proxy/internal/server/listener"
	"proxy/internal/server/pack"
)

type Server struct {
	extListener       *listener.ExternalListener
	intListener       *listener.InternalListener
	chanMsgToInternal chan pack.ChanProxyMessageToInternal
	chanMsgToExternal chan pack.ChanProxyMessageToExternal
}

func NewServer(externalPort string, advertisingAgentPort string) *Server {
	chanMsgToInternal := make(chan pack.ChanProxyMessageToInternal)
	chanMsgToExternal := make(chan pack.ChanProxyMessageToExternal)
	s := &Server{
		chanMsgToInternal: chanMsgToInternal,
		intListener:       listener.NewInternalListener(advertisingAgentPort, chanMsgToInternal),
		extListener:       listener.NewExternalListener(externalPort, chanMsgToInternal, chanMsgToExternal),
	}
	return s
}

func (s *Server) Start() {
	s.extListener.Run()
	s.intListener.Run()
}
