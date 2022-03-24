package enum

type ExternalToInternalMessageType string

const (
	MessageExternalToInternalMessageType         ExternalToInternalMessageType = "msg"
	CloseConnectionExternalToInternalMessageType ExternalToInternalMessageType = "cc"
)
