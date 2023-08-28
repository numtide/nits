package agent

import (
	"github.com/numtide/nits/internal/cmd"
	"github.com/numtide/nits/pkg/nats"
)

var Cmd struct {
	Nats    nats.NatsOptions `embed:"" prefix:"nats-"`
	Logging cmd.LogOptions   `embed:"" prefix:"log-"`

	Run  runCmd  `cmd:"" help:"Run an agent." default:"1"`
	Nkey nkeyCmd `cmd:"" help:"Produce a User NKey from an ed25519 key"`
}
