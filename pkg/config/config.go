package config

const (
	DefaultNatsURL = "ns://127.0.0.1:4222"
)

var DefaultNatsConfig = &Nats{
	Url: DefaultNatsURL,
}

type Nats struct {
	Url  string
	Jwt  string
	Seed string
}
