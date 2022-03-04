package server

import (
	"proxy/internal/server/listener"
	"proxy/internal/server/pack"
)

type Server struct {
	extListener       *listener.ExternalListener
	chanMsgToInternal chan pack.ChanProxyMessageToInternal
}

func NewServer(externalPort string) *Server {
	chanMsgToInternal := make(chan pack.ChanProxyMessageToInternal)
	s := &Server{
		chanMsgToInternal: chanMsgToInternal,
		extListener:       listener.NewExternalListener(externalPort, chanMsgToInternal),
	}
	return s
}

func (s *Server) Start() {
	s.extListener.Run()
}
