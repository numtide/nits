package server

import (
	"github.com/numtide/nits/internal/cmd"
	"github.com/numtide/nits/pkg/nutil"
)

var Cmd struct {
	Nats    nutil.NatsOptions `embed:"" prefix:"nats-"`
	Logging cmd.LogOptions    `embed:"" prefix:"log-"`
	Cache   cmd.CacheOptions  `embed:""`

	Run runCmd `cmd:"" default:"1"`
}
