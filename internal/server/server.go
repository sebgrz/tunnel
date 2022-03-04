package server

import "proxy/internal/server/listener"

type Server struct {
	extListener *listener.ExternalListener
}

func NewServer(extListener *listener.ExternalListener) *Server {
	s := &Server{extListener}
	return s
}

func (s *Server) Start() {
	s.extListener.Run()
}
