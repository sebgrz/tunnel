package inter

type SendConnection interface {
	Send(msgBytes []byte) error
}

type ListenConnection interface {
	Listen()
}

type ListenSendConnection interface {
	SendConnection
	ListenConnection
}
