package server

import (
	"github.com/numtide/nits/internal/cmd"
	"github.com/numtide/nits/pkg/nats"
)

var Cmd struct {
	Nats    nats.NatsOptions `embed:"" prefix:"nats-"`
	Logging cmd.LogOptions   `embed:"" prefix:"log-"`
	Cache   cmd.CacheOptions `embed:""`

	Run runCmd `cmd:"" default:"1"`
}
