package inter

type SendConnection interface {
	Send(externalConnectionID string, msgBytes []byte) error
}

type ListenConnection interface {
	Listen()
}

type ListenSendConnection interface {
	SendConnection
	ListenConnection
}
