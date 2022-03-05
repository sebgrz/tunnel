package inter

type SetHost interface {
	SetHost(host string)
}
type GetHost interface {
	GetHost() string
}

type Host interface {
	SetHost
	GetHost
}
