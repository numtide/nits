package guvnor

import (
	"github.com/numtide/nits/internal/cmd"
	"github.com/numtide/nits/internal/cmd/cache"
)

var Cmd struct {
	Nats    cmd.NatsOptions `embed:"" prefix:"nats-"`
	Logging cmd.LogOptions  `embed:"" prefix:"log-"`
	Cache   cache.Options   `embed:"" prefix:"cache-"`

	Run runCmd `cmd:"" default:"1"`
}
