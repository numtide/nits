package config

import "os"

const (
	DefaultNatsURL = "ns://127.0.0.1:4222"
)

var DefaultNatsConfig = &Nats{
	Url: DefaultNatsURL,
}

type Nats struct {
	Url         string
	Jwt         string
	Seed        string
	JwtFile     *os.File
	HostKeyFile *os.File
}
