package inter

type GetHost interface {
	GetHost() string
}

type Host interface {
	GetHost
}
