package server

import (
	"github.com/numtide/nits/internal/cmd"
)

var Cmd struct {
	Nats    cmd.NatsOptions  `embed:"" prefix:"nats-"`
	Logging cmd.LogOptions   `embed:"" prefix:"log-"`
	Cache   cmd.CacheOptions `embed:"" prefix:"cache-"`

	Run runCmd `cmd:"" default:"1"`
}
