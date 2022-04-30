package key

const (
	ExternalConnectionIDKey = "ec_id"
	// OriginHostname full hostname with subdoamin. NOT wildcard
	OriginHostname = "hsn"
	MessageTypeBytesHeader  = "msg_type"
)

const (
	CloseExternalPersistentConnectionMessageType = "cpc"
	CloseInternalConnectionMessageType           = "cic"
)

const (
	ConnectionHTTPHeader = "Connection"
	UpgradeHTTPHeader    = "Upgrade"
)
